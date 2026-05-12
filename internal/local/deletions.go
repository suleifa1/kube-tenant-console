package local

import (
	"fmt"

	"kube-tenant-console/internal/domain"
)

func (s *Store) DeleteTenant(id string) error {
	return s.Update(func(state *domain.State) error {
		tenant, found := FindTenant(state, id)
		if !found {
			return fmt.Errorf("tenant not found")
		}
		RemoveTenantState(state, tenant)
		state.Audit = append(state.Audit, Audit("tenant.deleted", fmt.Sprintf("deleted tenant %s from local state", tenant.Name)))
		return nil
	})
}

func (s *Store) DeleteNamespace(id string) error {
	return s.Update(func(state *domain.State) error {
		ns, found := FindNamespace(state, id)
		if !found {
			return fmt.Errorf("namespace not found")
		}
		RemoveNamespaceState(state, ns)
		state.Audit = append(state.Audit, Audit("namespace.deleted", fmt.Sprintf("deleted namespace %s from local state", ns.Name)))
		return nil
	})
}

func (s *Store) DeleteRole(id string) error {
	return s.Update(func(state *domain.State) error {
		role, found := FindRole(state, id)
		if !found {
			return fmt.Errorf("role not found")
		}
		RemoveRoleState(state, role)
		state.Audit = append(state.Audit, Audit("role.deleted", fmt.Sprintf("deleted role %s from local state", role.Name)))
		return nil
	})
}

func (s *Store) DeleteServiceAccount(id string) error {
	return s.Update(func(state *domain.State) error {
		serviceAccount, found := FindServiceAccount(state, id)
		if !found {
			return fmt.Errorf("service account not found")
		}
		RemoveServiceAccountState(state, serviceAccount)
		state.Audit = append(state.Audit, Audit("serviceaccount.deleted", fmt.Sprintf("deleted service account %s from local state", serviceAccount.Name)))
		return nil
	})
}

func (s *Store) DeleteAssignment(id string) error {
	return s.Update(func(state *domain.State) error {
		assignment, found := FindAssignment(state, id)
		if !found {
			return fmt.Errorf("assignment not found")
		}
		state.Assignments = FilterSlice(state.Assignments, func(item domain.Assignment) bool {
			return item.ID != assignment.ID
		})
		state.Audit = append(state.Audit, Audit("assignment.deleted", fmt.Sprintf("deleted binding for %s from local state", assignment.SubjectName)))
		return nil
	})
}

func (s *Store) DeleteKubeconfig(id string) error {
	return s.Update(func(state *domain.State) error {
		issue, found := FindKubeconfig(state, id)
		if !found {
			return fmt.Errorf("kubeconfig request not found")
		}
		state.Kubeconfigs = FilterSlice(state.Kubeconfigs, func(item domain.KubeconfigIssue) bool {
			return item.ID != issue.ID
		})
		state.Audit = append(state.Audit, Audit("kubeconfig.deleted", fmt.Sprintf("deleted kubeconfig request %s from local state", issue.Name)))
		return nil
	})
}

func RemoveTenantState(state *domain.State, tenant domain.Tenant) {
	namespaceIDs := make(map[string]bool)
	for _, ns := range state.Namespaces {
		if ns.TenantID == tenant.ID {
			namespaceIDs[ns.ID] = true
		}
	}

	roleIDs := make(map[string]bool)
	for _, role := range state.Roles {
		if role.TenantID == tenant.ID {
			roleIDs[role.ID] = true
		}
	}

	state.Tenants = FilterSlice(state.Tenants, func(item domain.Tenant) bool {
		return item.ID != tenant.ID
	})
	state.Namespaces = FilterSlice(state.Namespaces, func(item domain.TenantNamespace) bool {
		return item.TenantID != tenant.ID
	})
	state.Roles = FilterSlice(state.Roles, func(item domain.RoleTemplate) bool {
		return item.TenantID != tenant.ID
	})
	state.ServiceAccounts = FilterSlice(state.ServiceAccounts, func(item domain.TenantServiceAccount) bool {
		return item.TenantID != tenant.ID && !namespaceIDs[item.NamespaceID]
	})
	state.Assignments = FilterSlice(state.Assignments, func(item domain.Assignment) bool {
		return item.TenantID != tenant.ID && !namespaceIDs[item.NamespaceID] && !roleIDs[item.RoleID]
	})
	state.Kubeconfigs = FilterSlice(state.Kubeconfigs, func(item domain.KubeconfigIssue) bool {
		return item.TenantID != tenant.ID && !namespaceIDs[item.NamespaceID]
	})
}

func RemoveNamespaceState(state *domain.State, ns domain.TenantNamespace) {
	state.Namespaces = FilterSlice(state.Namespaces, func(item domain.TenantNamespace) bool {
		return item.ID != ns.ID
	})
	state.ServiceAccounts = FilterSlice(state.ServiceAccounts, func(item domain.TenantServiceAccount) bool {
		return item.NamespaceID != ns.ID
	})
	state.Assignments = FilterSlice(state.Assignments, func(item domain.Assignment) bool {
		return item.NamespaceID != ns.ID
	})
	state.Kubeconfigs = FilterSlice(state.Kubeconfigs, func(item domain.KubeconfigIssue) bool {
		return item.NamespaceID != ns.ID
	})
}

func RemoveRoleState(state *domain.State, role domain.RoleTemplate) {
	state.Roles = FilterSlice(state.Roles, func(item domain.RoleTemplate) bool {
		return item.ID != role.ID
	})
	state.Assignments = FilterSlice(state.Assignments, func(item domain.Assignment) bool {
		return item.RoleID != role.ID
	})
}

func RemoveServiceAccountState(state *domain.State, serviceAccount domain.TenantServiceAccount) {
	state.ServiceAccounts = FilterSlice(state.ServiceAccounts, func(item domain.TenantServiceAccount) bool {
		return item.ID != serviceAccount.ID
	})
	state.Assignments = FilterSlice(state.Assignments, func(item domain.Assignment) bool {
		return item.SubjectKind != "ServiceAccount" ||
			item.NamespaceID != serviceAccount.NamespaceID ||
			item.SubjectName != serviceAccount.Name
	})
	state.Kubeconfigs = FilterSlice(state.Kubeconfigs, func(item domain.KubeconfigIssue) bool {
		return item.NamespaceID != serviceAccount.NamespaceID || item.Name != serviceAccount.Name
	})
}
