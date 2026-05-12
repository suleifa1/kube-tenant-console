package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"kube-tenant-console/internal/domain"
	"kube-tenant-console/internal/kube"
	"kube-tenant-console/internal/local"
	"net/http"
	"strings"
	"time"
)

func NewServer(store *local.Store, cfg domain.Config, kubeClient ...*kube.KubeClient) *Server {
	var kube *KubeClient
	if len(kubeClient) > 0 {
		kube = kubeClient[0]
	}
	return &Server{store: store, cfg: cfg, kube: kube}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	s.registerApiRoutes(mux)
	s.registerStaticRoutes(mux)

	return requestLogger(mux)
}

func (s *Server) registerApiRoutes(mux *http.ServeMux) {
	s.registerApiGETRoutes(mux)
	s.registerApiPOSTRoutes(mux)
	s.registerApiDELETERoutes(mux)
}

func (s *Server) deleteClusterResource(w http.ResponseWriter, r *http.Request, deleteFunc func(context.Context, *KubeClient, string, string) error) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("namespace and name are required"))
		return
	}

	ctx, cancel := kubeContext(r)
	defer cancel()
	if err := deleteFunc(ctx, kube, namespace, name); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "delete-requested", "namespace": namespace, "name": name})
}

func (s *Server) requireKube(w http.ResponseWriter) (*KubeClient, bool) {
	if s.kube == nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("kubernetes client is not configured"))
		return nil, false
	}
	return s.kube, true
}

func kubeContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 30*time.Second)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func joinDocs(docs ...string) string {
	return strings.Join(docs, "---\n") + "\n"
}
