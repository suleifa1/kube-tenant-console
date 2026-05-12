package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/client-go/rest"
)

func TestPreviewNamespaceEndpoint(t *testing.T) {
	ns := testTenantNamespace()
	server := NewServer(newTestStore(t, State{
		Tenants:    []Tenant{{ID: ns.TenantID, Name: "team-a", NamespacePrefix: "team-a"}},
		Namespaces: []TenantNamespace{ns},
	}), Config{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/namespaces/"+ns.ID+"/yaml", nil)

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodePreviewResponse(t, response)
	t.Logf("rendered namespace YAML:\n%s", body.YAML)
	for _, want := range []string{
		"kind: Namespace",
		"kind: ResourceQuota",
		"kind: LimitRange",
		"name: team-a-dev",
		"name: tenant-quota",
		"name: tenant-defaults",
		"lrbac.dev/tenant-id: tenant-team-a",
	} {
		if !strings.Contains(body.YAML, want) {
			t.Fatalf("preview YAML missing %q:\n%s", want, body.YAML)
		}
	}
}

func TestPreviewRoleEndpoint(t *testing.T) {
	ns := testTenantNamespace()
	role := testRoleTemplate(ns.TenantID)
	server := NewServer(newTestStore(t, State{
		Tenants:    []Tenant{{ID: ns.TenantID, Name: "team-a", NamespacePrefix: "team-a"}},
		Namespaces: []TenantNamespace{ns},
		Roles:      []RoleTemplate{role},
	}), Config{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/namespaces/"+ns.ID+"/roles/"+role.ID+"/yaml", nil)

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodePreviewResponse(t, response)
	for _, want := range []string{
		"apiVersion: rbac.authorization.k8s.io/v1",
		"kind: Role",
		"name: deployer",
		"namespace: team-a-dev",
		"resources:",
		"verbs:",
	} {
		if !strings.Contains(body.YAML, want) {
			t.Fatalf("preview YAML missing %q:\n%s", want, body.YAML)
		}
	}
}

func TestPreviewKubeconfigEndpoint(t *testing.T) {
	ns := testTenantNamespace()
	tenant := Tenant{ID: ns.TenantID, Name: "team-a", NamespacePrefix: "team-a"}
	serviceAccount := TenantServiceAccount{
		ID:          "sa-developer",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		Name:        "developer",
	}
	issue := KubeconfigIssue{
		ID:          "kc-developer",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		Name:        serviceAccount.Name,
		TTLHours:    24,
	}
	server := NewServer(newTestStore(t, State{
		Tenants:         []Tenant{tenant},
		Namespaces:      []TenantNamespace{ns},
		ServiceAccounts: []TenantServiceAccount{serviceAccount},
		Kubeconfigs:     []KubeconfigIssue{issue},
	}), Config{}, &KubeClient{
		Config: &rest.Config{
			Host: "https://127.0.0.1:6443",
			TLSClientConfig: rest.TLSClientConfig{
				CAData: []byte("ca-data"),
			},
		},
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/kubeconfigs/"+issue.ID+"/yaml", nil)

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodePreviewResponse(t, response)
	for _, want := range []string{
		"kind: Config",
		"server: https://127.0.0.1:6443",
		"namespace: team-a-dev",
		"token: TOKEN_WILL_BE_ISSUED_BY_TOKENREQUEST",
	} {
		if !strings.Contains(body.YAML, want) {
			t.Fatalf("preview YAML missing %q:\n%s", want, body.YAML)
		}
	}
}

func testRoleTemplate(tenantID string) RoleTemplate {
	return RoleTemplate{
		ID:       "role-deployer",
		TenantID: tenantID,
		Name:     "deployer",
		Scope:    "namespace",
		Rules: []RoleRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
			},
		},
	}
}

func decodePreviewResponse(t *testing.T, response *httptest.ResponseRecorder) struct {
	YAML string `json:"yaml"`
} {
	t.Helper()

	var body struct {
		YAML string `json:"yaml"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode preview response: %v; body=%s", err, response.Body.String())
	}
	if body.YAML == "" {
		t.Fatalf("preview response has empty yaml: %s", response.Body.String())
	}
	return body
}
