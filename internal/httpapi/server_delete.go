package httpapi

import (
	"context"
	"fmt"
	"net/http"
)

func (s *Server) registerApiDELETERoutes(mux *http.ServeMux) {
	mux.HandleFunc("DELETE /api/tenants/{id}", s.handleDeleteTenant)
	mux.HandleFunc("DELETE /api/namespaces/{id}", s.handleDeleteNamespace)
	mux.HandleFunc("DELETE /api/roles/{id}", s.handleDeleteRole)
	mux.HandleFunc("DELETE /api/serviceaccounts/{id}", s.handleDeleteServiceAccount)
	mux.HandleFunc("DELETE /api/assignments/{id}", s.handleDeleteAssignment)
	mux.HandleFunc("DELETE /api/kubeconfigs/{id}", s.handleDeleteKubeconfig)
	mux.HandleFunc("DELETE /api/cluster/namespaces/{name}", s.handleDeleteClusterNamespace)
	mux.HandleFunc("DELETE /api/cluster/namespaces/{namespace}/serviceaccounts/{name}", s.handleDeleteClusterServiceAccount)
	mux.HandleFunc("DELETE /api/cluster/namespaces/{namespace}/roles/{name}", s.handleDeleteClusterRole)
	mux.HandleFunc("DELETE /api/cluster/namespaces/{namespace}/rolebindings/{name}", s.handleDeleteClusterRoleBinding)
	mux.HandleFunc("DELETE /api/cluster/namespaces/{namespace}/resourcequotas/{name}", s.handleDeleteClusterResourceQuota)
	mux.HandleFunc("DELETE /api/cluster/namespaces/{namespace}/limitranges/{name}", s.handleDeleteClusterLimitRange)
}

func (s *Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("tenant id is required"))
		return
	}

	if err := s.store.DeleteTenant(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) handleDeleteNamespace(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("namespace id is required"))
		return
	}

	if err := s.store.DeleteNamespace(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("role id is required"))
		return
	}

	if err := s.store.DeleteRole(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) handleDeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("service account id is required"))
		return
	}

	if err := s.store.DeleteServiceAccount(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) handleDeleteAssignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("assignment id is required"))
		return
	}

	if err := s.store.DeleteAssignment(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) handleDeleteKubeconfig(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("kubeconfig request id is required"))
		return
	}

	if err := s.store.DeleteKubeconfig(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) handleDeleteClusterNamespace(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("namespace name is required"))
		return
	}
	ctx, cancel := kubeContext(r)
	defer cancel()
	if err := kube.DeleteNamespace(ctx, name, ManagedSelectors()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "delete-requested", "namespace": name})
}

func (s *Server) handleDeleteClusterServiceAccount(w http.ResponseWriter, r *http.Request) {
	s.deleteClusterResource(w, r, func(ctx context.Context, kube *KubeClient, namespace, name string) error {
		return kube.DeleteServiceAccount(ctx, namespace, name, ManagedSelectors())
	})
}

func (s *Server) handleDeleteClusterRole(w http.ResponseWriter, r *http.Request) {
	s.deleteClusterResource(w, r, func(ctx context.Context, kube *KubeClient, namespace, name string) error {
		return kube.DeleteRole(ctx, namespace, name, ManagedSelectors())
	})
}

func (s *Server) handleDeleteClusterRoleBinding(w http.ResponseWriter, r *http.Request) {
	s.deleteClusterResource(w, r, func(ctx context.Context, kube *KubeClient, namespace, name string) error {
		return kube.DeleteRoleBinding(ctx, namespace, name, ManagedSelectors())
	})
}

func (s *Server) handleDeleteClusterResourceQuota(w http.ResponseWriter, r *http.Request) {
	s.deleteClusterResource(w, r, func(ctx context.Context, kube *KubeClient, namespace, name string) error {
		return kube.DeleteResourceQuota(ctx, namespace, name, ManagedSelectors())
	})
}

func (s *Server) handleDeleteClusterLimitRange(w http.ResponseWriter, r *http.Request) {
	s.deleteClusterResource(w, r, func(ctx context.Context, kube *KubeClient, namespace, name string) error {
		return kube.DeleteLimitRange(ctx, namespace, name, ManagedSelectors())
	})
}
