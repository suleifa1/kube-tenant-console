package local

import (
	"fmt"
	"time"

	"kube-tenant-console/internal/domain"
)

func (s *Store) CreateTenant(input domain.Tenant) (domain.Tenant, error) {
	if err := domain.ValidateName("tenant", input.Name); err != nil {
		return domain.Tenant{}, err
	}
	if input.NamespacePrefix == "" {
		input.NamespacePrefix = input.Name
	}
	if err := domain.ValidateName("namespace prefix", input.NamespacePrefix); err != nil {
		return domain.Tenant{}, err
	}
	input.ID = NewID("tenant")
	input.CreatedAt = time.Now().UTC()

	err := s.Update(func(state *domain.State) error {
		state.Tenants = append(state.Tenants, input)
		state.Audit = append(state.Audit, Audit("tenant.created", fmt.Sprintf("created tenant %s", input.Name)))
		return nil
	})
	return input, err
}

func (s *Store) CreateNamespace(input domain.TenantNamespace) (domain.TenantNamespace, error) {
	if err := domain.ValidateName("namespace", input.Name); err != nil {
		return domain.TenantNamespace{}, err
	}
	input.ID = NewID("ns")
	input.CreatedAt = time.Now().UTC()
	input.Labels = map[string]string{"tenant": input.TenantID}
	input.Quota = domain.DefaultQuota(input.Quota)
	input.LimitRange = domain.DefaultLimitRange(input.LimitRange)

	err := s.Update(func(state *domain.State) error {
		if !TenantExists(state, input.TenantID) {
			return fmt.Errorf("tenant not found")
		}
		state.Namespaces = append(state.Namespaces, input)
		state.Audit = append(state.Audit, Audit("namespace.created", fmt.Sprintf("created namespace %s", input.Name)))
		return nil
	})
	return input, err
}

func (s *Store) CreateRole(input domain.RoleTemplate, cfg domain.Config) (domain.RoleTemplate, error) {
	if input.Scope == "" {
		input.Scope = "namespace"
	}
	if err := domain.ValidateRole(input, cfg); err != nil {
		return domain.RoleTemplate{}, err
	}
	input.ID = NewID("role")
	input.CreatedAt = time.Now().UTC()

	err := s.Update(func(state *domain.State) error {
		if !TenantExists(state, input.TenantID) {
			return fmt.Errorf("tenant not found")
		}
		state.Roles = append(state.Roles, input)
		state.Audit = append(state.Audit, Audit("role.created", fmt.Sprintf("created role %s", input.Name)))
		return nil
	})
	return input, err
}

func (s *Store) CreateAssignment(input domain.Assignment) (domain.Assignment, error) {
	input.ID = NewID("bind")
	input.CreatedAt = time.Now().UTC()
	if input.SubjectKind == "" {
		input.SubjectKind = "User"
	}
	if !domain.ValidateSubjectKind(input.SubjectKind) {
		return domain.Assignment{}, fmt.Errorf("subject kind must be User, Group, or ServiceAccount")
	}
	if input.SubjectName == "" {
		return domain.Assignment{}, fmt.Errorf("subject name is required")
	}

	err := s.Update(func(state *domain.State) error {
		role, ok := FindRole(state, input.RoleID)
		if !ok {
			return fmt.Errorf("role not found")
		}
		ns, ok := FindNamespace(state, input.NamespaceID)
		if !ok {
			return fmt.Errorf("namespace not found")
		}
		if role.TenantID != ns.TenantID {
			return fmt.Errorf("role does not belong to namespace tenant")
		}
		input.TenantID = ns.TenantID
		if domain.IsAssignmentExists(state, input) {
			return fmt.Errorf("assignment already exists.")
		}
		if input.SubjectKind == "ServiceAccount" && !domain.IsServiceAccountExists(state, input.NamespaceID, input.SubjectName) {
			return fmt.Errorf("service account for binding not found")
		}

		state.Assignments = append(state.Assignments, input)
		state.Audit = append(state.Audit, Audit("assignment.created", fmt.Sprintf("bound %s to %s", input.SubjectName, role.Name)))
		return nil
	})
	return input, err
}

func (s *Store) CreateKubeconfig(input domain.KubeconfigIssue) (domain.KubeconfigIssue, error) {
	if err := domain.ValidateName("service account", input.Name); err != nil {
		return domain.KubeconfigIssue{}, err
	}
	if input.TTLHours < 0 {
		return domain.KubeconfigIssue{}, fmt.Errorf("ttl hours must be positive")
	}
	if input.TTLHours == 0 {
		input.TTLHours = 24
	}
	input.ID = NewID("kc")
	input.CreatedAt = time.Now().UTC()

	err := s.Update(func(state *domain.State) error {
		ns, ok := FindNamespace(state, input.NamespaceID)
		if !ok {
			return fmt.Errorf("namespace not found")
		}
		if !domain.IsServiceAccountExists(state, input.NamespaceID, input.Name) {
			return fmt.Errorf("service account for kubeconfig not found")
		}
		input.TenantID = ns.TenantID
		state.Kubeconfigs = append(state.Kubeconfigs, input)
		state.Audit = append(state.Audit, Audit("kubeconfig.requested", fmt.Sprintf("requested kubeconfig %s", input.Name)))
		return nil
	})
	return input, err
}

func (s *Store) CreateServiceAccount(input domain.TenantServiceAccount) (domain.TenantServiceAccount, error) {
	if err := domain.ValidateName("service account", input.Name); err != nil {
		return domain.TenantServiceAccount{}, err
	}
	input.ID = NewID("sa")
	if input.NamespaceID == "" {
		return domain.TenantServiceAccount{}, fmt.Errorf("Missing namespace ID.")
	}

	err := s.Update(func(state *domain.State) error {
		ns, found := FindNamespace(state, input.NamespaceID)
		if !found {
			return fmt.Errorf("The ID is not associated with any namespace.")
		}
		input.TenantID = ns.TenantID
		input.CreatedAt = time.Now().UTC()
		if domain.IsServiceAccountExists(state, input.NamespaceID, input.Name) {
			return fmt.Errorf("The service account is already exists.")
		}
		state.ServiceAccounts = append(state.ServiceAccounts, input)
		state.Audit = append(state.Audit, Audit("serviceaccount.created", fmt.Sprintf("created service account %s", input.Name)))
		return nil
	})
	return input, err
}
