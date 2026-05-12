package kube

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"kube-tenant-console/internal/domain"
)

func TestBuildNamespace(t *testing.T) {
	ns := testTenantNamespace()

	obj := BuildNamespace(ns)

	if obj.APIVersion != "v1" {
		t.Fatalf("APIVersion = %q, want v1", obj.APIVersion)
	}
	if obj.Kind != "Namespace" {
		t.Fatalf("Kind = %q, want Namespace", obj.Kind)
	}
	if obj.Name != ns.Name {
		t.Fatalf("Name = %q, want %q", obj.Name, ns.Name)
	}
	if obj.Labels[ManagedByLabel] != AppName {
		t.Fatalf("%s label = %q, want %q", ManagedByLabel, obj.Labels[ManagedByLabel], AppName)
	}
	if obj.Labels[TenantIDLabel] != ns.TenantID {
		t.Fatalf("%s label = %q, want %q", TenantIDLabel, obj.Labels[TenantIDLabel], ns.TenantID)
	}
}

func TestBuildResourceQuota(t *testing.T) {
	ns := testTenantNamespace()

	obj := BuildResourceQuota(ns)

	if obj.APIVersion != "v1" {
		t.Fatalf("APIVersion = %q, want v1", obj.APIVersion)
	}
	if obj.Kind != "ResourceQuota" {
		t.Fatalf("Kind = %q, want ResourceQuota", obj.Kind)
	}
	if obj.Name != "tenant-quota" {
		t.Fatalf("Name = %q, want tenant-quota", obj.Name)
	}
	if obj.Namespace != ns.Name {
		t.Fatalf("Namespace = %q, want %q", obj.Namespace, ns.Name)
	}

	assertQuantity(t, obj.Spec.Hard, corev1.ResourceRequestsCPU, ns.Quota.RequestsCPU)
	assertQuantity(t, obj.Spec.Hard, corev1.ResourceRequestsMemory, ns.Quota.RequestsMemory)
	assertQuantity(t, obj.Spec.Hard, corev1.ResourceLimitsCPU, ns.Quota.LimitsCPU)
	assertQuantity(t, obj.Spec.Hard, corev1.ResourceLimitsMemory, ns.Quota.LimitsMemory)
	assertQuantity(t, obj.Spec.Hard, corev1.ResourcePods, ns.Quota.Pods)
	assertQuantity(t, obj.Spec.Hard, corev1.ResourcePersistentVolumeClaims, ns.Quota.PVCs)
	assertQuantity(t, obj.Spec.Hard, corev1.ResourceRequestsStorage, ns.Quota.Storage)
}

func TestNamespaceToYAML(t *testing.T) {
	obj := BuildNamespace(testTenantNamespace())

	got := ObjectToYAML(obj)
	t.Logf("rendered namespace YAML:\n%s", got)

	for _, want := range []string{
		"apiVersion: v1",
		"kind: Namespace",
		"name: team-a-dev",
		"app.kubernetes.io/managed-by: lrbac",
		"lrbac.dev/tenant-id: tenant-team-a",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered YAML missing %q:\n%s", want, got)
		}
	}
}

func TestBuildKubeConfigNamespace(t *testing.T) {
	ns := testTenantNamespace()
	tenant := domain.Tenant{ID: ns.TenantID, Name: "team-a"}
	sa := domain.TenantServiceAccount{
		ID:          "sa-dev",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		Name:        "developer",
	}
	issue := domain.KubeconfigIssue{
		ID:          "kc-dev",
		TenantID:    ns.TenantID,
		NamespaceID: ns.ID,
		Name:        sa.Name,
		TTLHours:    24,
	}

	cfg := BuildKubeConfigNamespace(ns, sa, issue, tenant, "https://127.0.0.1:6443", []byte("ca-data"), "token-value")
	got, err := KubeConfigToYAML(cfg)
	if err != nil {
		t.Fatalf("render kubeconfig: %v", err)
	}
	t.Logf("rendered kubeconfig YAML:\n%s", got)

	for _, want := range []string{
		"apiVersion: v1",
		"kind: Config",
		"server: https://127.0.0.1:6443",
		"certificate-authority-data:",
		"name: team-a-dev",
		"namespace: team-a-dev",
		"token: token-value",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered kubeconfig missing %q:\n%s", want, got)
		}
	}
}

func testTenantNamespace() domain.TenantNamespace {
	return domain.TenantNamespace{
		ID:       "ns-team-a-dev",
		TenantID: "tenant-team-a",
		Name:     "team-a-dev",
		Quota: domain.QuotaSpec{
			RequestsCPU:    "4",
			RequestsMemory: "8Gi",
			LimitsCPU:      "8",
			LimitsMemory:   "16Gi",
			Pods:           "30",
			PVCs:           "10",
			Storage:        "100Gi",
		},
		LimitRange: domain.LimitRangeSpec{
			DefaultCPU:    "500m",
			DefaultMemory: "512Mi",
			RequestCPU:    "100m",
			RequestMemory: "128Mi",
			MaxCPU:        "2",
			MaxMemory:     "4Gi",
		},
	}
}

func assertQuantity(t *testing.T, list corev1.ResourceList, name corev1.ResourceName, want string) {
	t.Helper()

	got, ok := list[name]
	if !ok {
		t.Fatalf("resource %s is missing", name)
	}

	wantQuantity := resource.MustParse(want)
	if got.Cmp(wantQuantity) != 0 {
		t.Fatalf("resource %s = %s, want %s", name, got.String(), wantQuantity.String())
	}
}
