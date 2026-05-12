package httpapi

import (
	"fmt"
	"kube-tenant-console/internal/local"
	"net/http"
	"time"
)

func (s *Server) registerApiPOSTRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/tenants", s.handleCreateTenant)
	mux.HandleFunc("POST /api/namespaces", s.handleCreateNamespace)
	mux.HandleFunc("POST /api/roles", s.handleCreateRole)
	mux.HandleFunc("POST /api/serviceaccounts", s.handleCreateServiceAccount)
	mux.HandleFunc("POST /api/assignments", s.handleCreateAssignment)
	mux.HandleFunc("POST /api/kubeconfigs", s.handleCreateKubeconfig)
	mux.HandleFunc("POST /api/kubeconfigs/{id}/token", s.handleIssueKubeconfigToken)
	mux.HandleFunc("POST /api/cluster/namespaces/{id}", s.handleEnsureClusterNamespace)
	mux.HandleFunc("POST /api/cluster/namespaces/{id}/roles/{roleId}", s.handleEnsureClusterRole)
	mux.HandleFunc("POST /api/cluster/namespaces/{id}/serviceaccounts/{saId}", s.handleEnsureClusterServiceAccount)
	mux.HandleFunc("POST /api/cluster/assignments/{id}", s.handleEnsureClusterAssignment)
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var input Tenant
	if !decodeJSON(w, r, &input) {
		return
	}
	tenant, err := s.store.CreateTenant(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, tenant)
}

func (s *Server) handleCreateNamespace(w http.ResponseWriter, r *http.Request) {
	var input TenantNamespace
	if !decodeJSON(w, r, &input) {
		return
	}
	ns, err := s.store.CreateNamespace(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, ns)
}

func (s *Server) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	var input RoleTemplate
	if !decodeJSON(w, r, &input) {
		return
	}
	role, err := s.store.CreateRole(input, s.cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

func (s *Server) handleCreateAssignment(w http.ResponseWriter, r *http.Request) {
	var input Assignment
	if !decodeJSON(w, r, &input) {
		return
	}
	assignment, err := s.store.CreateAssignment(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, assignment)
}

func (s *Server) handleCreateKubeconfig(w http.ResponseWriter, r *http.Request) {
	var input KubeconfigIssue
	if !decodeJSON(w, r, &input) {
		return
	}
	issue, err := s.store.CreateKubeconfig(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, issue)
}

func (s *Server) handleCreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var input TenantServiceAccount
	if !decodeJSON(w, r, &input) {
		return
	}
	serviceAccount, err := s.store.CreateServiceAccount(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, serviceAccount)
}

func (s *Server) handleIssueKubeconfigToken(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	issueID := r.PathValue("id")
	state := s.store.Snapshot()
	issue, ns, sa, tenant, err := local.ResolveKubeconfigIssue(&state, issueID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if issue.TTLHours == 0 {
		issue.TTLHours = 24
	}

	ctx, cancel := kubeContext(r)
	defer cancel()
	tokenRequest := BuildServiceAccountTokenRequest(issue.TTLHours)
	issuedToken, err := kube.IssueKubeconfigToken(ctx, ns.Name, sa.Name, tokenRequest)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if issuedToken.Status.Token == "" {
		writeError(w, http.StatusBadGateway, fmt.Errorf("token request returned empty token"))
		return
	}

	yaml, err := kube.RenderKubeconfig(issue, ns, sa, tenant, issuedToken.Status.Token)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	issuedAt := time.Now().UTC()
	expiresAt := issuedToken.Status.ExpirationTimestamp.Time.UTC()
	err = s.store.Update(func(state *State) error {
		for i := range state.Kubeconfigs {
			if state.Kubeconfigs[i].ID == issueID {
				state.Kubeconfigs[i].IssuedAt = &issuedAt
				state.Kubeconfigs[i].ExpiresAt = &expiresAt
				state.Kubeconfigs[i].Kubeconfig = ""
				issue = state.Kubeconfigs[i]
				state.Audit = append(state.Audit, local.Audit("kubeconfig.issued", fmt.Sprintf("issued kubeconfig %s", issue.Name)))
				return nil
			}
		}
		return fmt.Errorf("kubeconfig request not found")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, KubeconfigIssueResult{Issue: issue, YAML: yaml})
}

func (s *Server) handleEnsureClusterNamespace(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	namespaceID := r.PathValue("id")
	state := s.store.Snapshot()
	ns, found := local.FindNamespace(&state, namespaceID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("namespace was not found"))
		return
	}

	ctx, cancel := kubeContext(r)
	defer cancel()
	if err := kube.EnsureNamespaceBundle(ctx, ns); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "created-or-existing", "namespace": ns.Name})
}

func (s *Server) handleEnsureClusterRole(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	namespaceID := r.PathValue("id")
	roleID := r.PathValue("roleId")
	state := s.store.Snapshot()
	ns, found := local.FindNamespace(&state, namespaceID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("namespace was not found"))
		return
	}
	role, found := local.FindRole(&state, roleID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("role was not found"))
		return
	}
	if role.TenantID != ns.TenantID {
		writeError(w, http.StatusBadRequest, fmt.Errorf("role does not belong to namespace tenant"))
		return
	}
	if role.Scope == "cluster" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("cluster role apply is not implemented yet"))
		return
	}

	ctx, cancel := kubeContext(r)
	defer cancel()
	if err := kube.EnsureRole(ctx, BuildRole(role, ns)); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "created-or-existing", "role": role.Name, "namespace": ns.Name})
}

func (s *Server) handleEnsureClusterServiceAccount(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	namespaceID := r.PathValue("id")
	serviceAccountID := r.PathValue("saId")
	state := s.store.Snapshot()
	ns, found := local.FindNamespace(&state, namespaceID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("namespace was not found"))
		return
	}
	serviceAccount, found := local.FindServiceAccount(&state, serviceAccountID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("service account was not found"))
		return
	}
	if serviceAccount.NamespaceID != ns.ID {
		writeError(w, http.StatusBadRequest, fmt.Errorf("service account does not belong to namespace"))
		return
	}

	ctx, cancel := kubeContext(r)
	defer cancel()
	if err := kube.EnsureServiceAccount(ctx, BuildServiceAccount(ns, serviceAccount)); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "created-or-existing", "serviceAccount": serviceAccount.Name, "namespace": ns.Name})
}

func (s *Server) handleEnsureClusterAssignment(w http.ResponseWriter, r *http.Request) {
	kube, ok := s.requireKube(w)
	if !ok {
		return
	}
	assignmentID := r.PathValue("id")
	state := s.store.Snapshot()
	assignment, found := local.FindAssignment(&state, assignmentID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("assignment was not found"))
		return
	}
	role, found := local.FindRole(&state, assignment.RoleID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("role for assignment was not found"))
		return
	}
	ns, found := local.FindNamespace(&state, assignment.NamespaceID)
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("namespace for assignment was not found"))
		return
	}
	if role.TenantID != ns.TenantID {
		writeError(w, http.StatusBadRequest, fmt.Errorf("role does not belong to namespace tenant"))
		return
	}

	ctx, cancel := kubeContext(r)
	defer cancel()
	if err := kube.EnsureRoleBinding(ctx, BuildRoleBind(assignment, role, ns)); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "created-or-existing", "roleBinding": RoleBindingName(assignment, role), "namespace": ns.Name})
}
