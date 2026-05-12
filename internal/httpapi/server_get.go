package httpapi

import (
	"errors"
	"fmt"
	"kube-tenant-console/internal/local"
	"net/http"
)

func (s *Server) registerApiGETRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("GET /api/cluster", s.handleClusterSnapshot)
	mux.HandleFunc("GET /api/namespaces/{id}/yaml", s.previewNamespace)
	mux.HandleFunc("GET /api/namespaces/{id}/roles/{roleId}/yaml", s.previewRole)
	mux.HandleFunc("GET /api/namespaces/{id}/serviceaccounts/{saId}/yaml", s.previewServiceAccount)
	mux.HandleFunc("GET /api/assignments/{id}/yaml", s.previewAssignment)
	mux.HandleFunc("GET /api/kubeconfigs/{id}/yaml", s.previewKubeconfig)
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot())
}

func (s *Server) handleClusterSnapshot(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	ctx, cancel := kubeContext(r)
	defer cancel()

	snapshot, err := kube.GetManagedClusterSnapshot(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) previewNamespace(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("Bad Request"))
		return
	}
	state := s.store.Snapshot()
	ns, found := local.FindNamespace(&state, id)
	if !found {
		writeError(w, http.StatusNotFound, errors.New("Namespace was not found"))
		return
	}
	objectNs := BuildNamespace(ns)
	objectRq := BuildResourceQuota(ns)
	objectLr := BuildLimitRange(ns)
	yamls := joinDocs(ObjectToYAML(objectNs), ObjectToYAML(objectRq), ObjectToYAML(objectLr))
	writeJSON(w, http.StatusOK, map[string]string{"yaml": yamls})
}

func (s *Server) previewRole(w http.ResponseWriter, r *http.Request) {
	nsId := r.PathValue("id")
	if nsId == "" {
		writeError(w, http.StatusBadRequest, errors.New("Bad request. Missing namespace id."))
		return
	}
	roleId := r.PathValue("roleId")
	if roleId == "" {
		writeError(w, http.StatusBadRequest, errors.New("Bad request. Missing role id."))
		return
	}
	state := s.store.Snapshot()
	ns, found := local.FindNamespace(&state, nsId)
	if !found {
		writeError(w, http.StatusNotFound, errors.New("Namespace was not found"))
		return
	}
	role, found := local.FindRole(&state, roleId)
	if !found {
		writeError(w, http.StatusNotFound, errors.New("Role was not found"))
		return
	}
	if role.TenantID != ns.TenantID {
		writeError(w, http.StatusBadRequest, errors.New("Role does not belong to namespace"))
		return
	}
	if role.Scope == "cluster" {
		writeError(w, http.StatusBadRequest, errors.New("Cluster role preview is not supported in namespace endpoint"))
		return
	}
	object := BuildRole(role, ns)
	yaml := ObjectToYAML(object)

	writeJSON(w, http.StatusOK, map[string]string{"yaml": yaml})
}

func (s *Server) previewServiceAccount(w http.ResponseWriter, r *http.Request) {
	state := s.store.Snapshot()

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("namespace id was not presented."))
		return
	}
	saID := r.PathValue("saId")
	if saID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("service account id was not presented."))
		return
	}

	ns, found := local.FindNamespace(&state, id)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("that namespace ID is not belong to any namespace"))
		return
	}
	sa, found := local.FindServiceAccount(&state, saID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("that service account id is not belong to any service account."))
		return
	}
	if sa.NamespaceID != ns.ID {
		writeError(w, http.StatusBadRequest, fmt.Errorf("service account does not belong to namespace"))
		return
	}

	object := BuildServiceAccount(ns, sa)
	out := ObjectToYAML(object)
	writeJSON(w, http.StatusOK, map[string]string{"yaml": out})
}

func (s *Server) previewAssignment(w http.ResponseWriter, r *http.Request) {
	state := s.store.Snapshot()

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("namespace id was not presented."))
		return
	}

	assignment, found := local.FindAssignment(&state, id)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("that assignment ID is not exists"))
		return
	}

	role, found := local.FindRole(&state, assignment.RoleID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("the role belongs to that assignment is not exists."))
		return
	}

	ns, found := local.FindNamespace(&state, assignment.NamespaceID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("the namespace belongs to that assignment is not exists."))
		return
	}

	object := BuildRoleBind(assignment, role, ns)
	out := ObjectToYAML(object)
	writeJSON(w, http.StatusOK, map[string]string{"yaml": out})
}

func (s *Server) previewKubeconfig(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	state := s.store.Snapshot()
	issue, ns, sa, tenant, err := local.ResolveKubeconfigIssue(&state, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	yaml, err := kube.RenderKubeconfig(issue, ns, sa, tenant, "")
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, KubeconfigIssueResult{Issue: issue, YAML: yaml})
}
