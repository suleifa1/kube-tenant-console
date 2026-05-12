package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteNamespaceEndpointCascadesLocalState(t *testing.T) {
	ns := testTenantNamespace()
	role := testRoleTemplate(ns.TenantID)
	serviceAccount := TenantServiceAccount{
		ID:          "sa-developer",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		Name:        "developer",
	}
	assignment := Assignment{
		ID:          "bind-developer",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		RoleID:      role.ID,
		SubjectKind: "ServiceAccount",
		SubjectName: serviceAccount.Name,
	}
	issue := KubeconfigIssue{
		ID:          "kc-developer",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		Name:        serviceAccount.Name,
		TTLHours:    24,
	}
	store := newTestStore(t, State{
		Tenants:         []Tenant{{ID: ns.TenantID, Name: "team-a", NamespacePrefix: "team-a"}},
		Namespaces:      []TenantNamespace{ns},
		Roles:           []RoleTemplate{role},
		ServiceAccounts: []TenantServiceAccount{serviceAccount},
		Assignments:     []Assignment{assignment},
		Kubeconfigs:     []KubeconfigIssue{issue},
	})
	server := NewServer(store, Config{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/namespaces/"+ns.ID, nil)

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	state := store.Snapshot()
	if len(state.Namespaces) != 0 {
		t.Fatalf("namespaces len = %d, want 0", len(state.Namespaces))
	}
	if len(state.ServiceAccounts) != 0 {
		t.Fatalf("service accounts len = %d, want 0", len(state.ServiceAccounts))
	}
	if len(state.Assignments) != 0 {
		t.Fatalf("assignments len = %d, want 0", len(state.Assignments))
	}
	if len(state.Kubeconfigs) != 0 {
		t.Fatalf("kubeconfigs len = %d, want 0", len(state.Kubeconfigs))
	}
	if len(state.Roles) != 1 {
		t.Fatalf("roles len = %d, want 1", len(state.Roles))
	}
}

func TestDeleteRoleEndpointCascadesAssignments(t *testing.T) {
	ns := testTenantNamespace()
	role := testRoleTemplate(ns.TenantID)
	store := newTestStore(t, State{
		Tenants:    []Tenant{{ID: ns.TenantID, Name: "team-a", NamespacePrefix: "team-a"}},
		Namespaces: []TenantNamespace{ns},
		Roles:      []RoleTemplate{role},
		Assignments: []Assignment{
			{ID: "bind-user", TenantID: ns.TenantID, NamespaceID: ns.ID, RoleID: role.ID, SubjectKind: "User", SubjectName: "alice@example.com"},
		},
	})
	server := NewServer(store, Config{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/roles/"+role.ID, nil)

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	state := store.Snapshot()
	if len(state.Roles) != 0 {
		t.Fatalf("roles len = %d, want 0", len(state.Roles))
	}
	if len(state.Assignments) != 0 {
		t.Fatalf("assignments len = %d, want 0", len(state.Assignments))
	}
	if len(state.Namespaces) != 1 {
		t.Fatalf("namespaces len = %d, want 1", len(state.Namespaces))
	}
}
