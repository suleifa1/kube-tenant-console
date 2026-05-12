package local

import (
	"fmt"
	"time"

	"kube-tenant-console/internal/domain"
)

func NewID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

func Audit(action string, summary string) domain.AuditEvent {
	return domain.AuditEvent{ID: NewID("audit"), Action: action, Summary: summary, CreatedAt: time.Now().UTC()}
}

func TenantExists(state *domain.State, id string) bool {
	for _, tenant := range state.Tenants {
		if tenant.ID == id {
			return true
		}
	}
	return false
}

func FindTenant(state *domain.State, id string) (domain.Tenant, bool) {
	for _, tenant := range state.Tenants {
		if tenant.ID == id {
			return tenant, true
		}
	}
	return domain.Tenant{}, false
}

func FindNamespace(state *domain.State, id string) (domain.TenantNamespace, bool) {
	for _, ns := range state.Namespaces {
		if ns.ID == id {
			return ns, true
		}
	}
	return domain.TenantNamespace{}, false
}

func FindRole(state *domain.State, id string) (domain.RoleTemplate, bool) {
	for _, role := range state.Roles {
		if role.ID == id {
			return role, true
		}
	}
	return domain.RoleTemplate{}, false
}

func FindAssignment(state *domain.State, id string) (domain.Assignment, bool) {
	for _, assignment := range state.Assignments {
		if assignment.ID == id {
			return assignment, true
		}
	}
	return domain.Assignment{}, false
}

func FindServiceAccount(state *domain.State, id string) (domain.TenantServiceAccount, bool) {
	for _, serviceAccount := range state.ServiceAccounts {
		if serviceAccount.ID == id {
			return serviceAccount, true
		}
	}
	return domain.TenantServiceAccount{}, false
}

func FindServiceAccountByName(state *domain.State, namespaceID string, name string) (domain.TenantServiceAccount, bool) {
	for _, serviceAccount := range state.ServiceAccounts {
		if serviceAccount.NamespaceID == namespaceID && serviceAccount.Name == name {
			return serviceAccount, true
		}
	}
	return domain.TenantServiceAccount{}, false
}

func FindKubeconfig(state *domain.State, id string) (domain.KubeconfigIssue, bool) {
	for _, issue := range state.Kubeconfigs {
		if issue.ID == id {
			return issue, true
		}
	}
	return domain.KubeconfigIssue{}, false
}

func ResolveKubeconfigIssue(state *domain.State, id string) (domain.KubeconfigIssue, domain.TenantNamespace, domain.TenantServiceAccount, domain.Tenant, error) {
	issue, found := FindKubeconfig(state, id)
	if !found {
		return domain.KubeconfigIssue{}, domain.TenantNamespace{}, domain.TenantServiceAccount{}, domain.Tenant{}, fmt.Errorf("kubeconfig request was not found")
	}
	ns, found := FindNamespace(state, issue.NamespaceID)
	if !found {
		return domain.KubeconfigIssue{}, domain.TenantNamespace{}, domain.TenantServiceAccount{}, domain.Tenant{}, fmt.Errorf("namespace for kubeconfig was not found")
	}
	sa, found := FindServiceAccountByName(state, issue.NamespaceID, issue.Name)
	if !found {
		return domain.KubeconfigIssue{}, domain.TenantNamespace{}, domain.TenantServiceAccount{}, domain.Tenant{}, fmt.Errorf("service account for kubeconfig was not found")
	}
	tenant, found := FindTenant(state, ns.TenantID)
	if !found {
		return domain.KubeconfigIssue{}, domain.TenantNamespace{}, domain.TenantServiceAccount{}, domain.Tenant{}, fmt.Errorf("tenant for kubeconfig was not found")
	}
	return issue, ns, sa, tenant, nil
}

func FilterSlice[T any](items []T, keep func(T) bool) []T {
	out := items[:0]
	for _, item := range items {
		if keep(item) {
			out = append(out, item)
		}
	}
	return out
}
