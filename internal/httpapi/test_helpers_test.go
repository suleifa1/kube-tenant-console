package httpapi

import (
	"path/filepath"
	"testing"

	"kube-tenant-console/internal/local"
)

func newTestStore(t *testing.T, state State) *Store {
	t.Helper()

	store, err := local.OpenStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := store.Update(func(current *State) error {
		*current = state
		return nil
	}); err != nil {
		t.Fatalf("seed test store: %v", err)
	}
	return store
}

func testTenantNamespace() TenantNamespace {
	return TenantNamespace{
		ID:       "ns-team-a-dev",
		TenantID: "tenant-team-a",
		Name:     "team-a-dev",
		Quota: QuotaSpec{
			RequestsCPU:    "4",
			RequestsMemory: "8Gi",
			LimitsCPU:      "8",
			LimitsMemory:   "16Gi",
			Pods:           "30",
			PVCs:           "10",
			Storage:        "100Gi",
		},
		LimitRange: LimitRangeSpec{
			DefaultCPU:    "500m",
			DefaultMemory: "512Mi",
			RequestCPU:    "100m",
			RequestMemory: "128Mi",
			MaxCPU:        "2",
			MaxMemory:     "4Gi",
		},
	}
}
