package local

import (
	"encoding/json"
	"errors"
	"kube-tenant-console/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	path  string
	mu    sync.Mutex
	state domain.State
}

func OpenStore(path string) (*Store, error) {
	store := &Store{path: path}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		store.state = emptyState()
		return store, store.saveLocked()
	}
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &store.state); err != nil {
			return nil, err
		}
	}
	normalizeState(&store.state)
	return store, nil
}

func (s *Store) Snapshot() domain.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := cloneState(s.state)
	normalizeState(&state)
	return state
}

func (s *Store) Update(fn func(*domain.State) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(&s.state); err != nil {
		return err
	}
	normalizeState(&s.state)
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	normalizeState(&s.state)
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func cloneState(state domain.State) domain.State {
	data, _ := json.Marshal(state)
	var out domain.State
	_ = json.Unmarshal(data, &out)
	return out
}

func emptyState() domain.State {
	state := domain.State{}
	normalizeState(&state)
	return state
}

func normalizeState(state *domain.State) {
	if state.Tenants == nil {
		state.Tenants = []domain.Tenant{}
	}
	if state.Namespaces == nil {
		state.Namespaces = []domain.TenantNamespace{}
	}
	if state.Roles == nil {
		state.Roles = []domain.RoleTemplate{}
	}
	if state.Assignments == nil {
		state.Assignments = []domain.Assignment{}
	}
	if state.Kubeconfigs == nil {
		state.Kubeconfigs = []domain.KubeconfigIssue{}
	}
	if state.Audit == nil {
		state.Audit = []domain.AuditEvent{}
	}
	if state.ServiceAccounts == nil {
		state.ServiceAccounts = []domain.TenantServiceAccount{}
	}
}
