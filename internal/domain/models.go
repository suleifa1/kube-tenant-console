package domain

import "time"

// Config is runtime configuration for the app process.
// It is not tenant data; it describes which high-risk features are enabled.
type Config struct {
	AllowClusterScope bool
}

// State is the whole local application database.
// In the current MVP it is stored as JSON on disk and represents what users
// created through the UI: tenants, namespaces, roles, bindings, kubeconfig
// requests, and audit events.
type State struct {
	Tenants         []Tenant               `json:"tenants"`
	Namespaces      []TenantNamespace      `json:"namespaces"`
	Roles           []RoleTemplate         `json:"roles"`
	Assignments     []Assignment           `json:"assignments"`
	Kubeconfigs     []KubeconfigIssue      `json:"kubeconfigs"`
	ServiceAccounts []TenantServiceAccount `json:"serviceaccounts"`
	Audit           []AuditEvent           `json:"audit"`
}

// Tenant is our product-level owner/group, not a native Kubernetes resource.
// A tenant usually represents a team, customer, project, or department and owns
// one or more Kubernetes namespaces.
type Tenant struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	NamespacePrefix string    `json:"namespacePrefix"`
	CreatedAt       time.Time `json:"createdAt"`
}

// TenantNamespace is a Kubernetes namespace managed for a tenant.
// It links our tenant model to a real namespace name and carries default quota
// and limit settings for that namespace.
type TenantNamespace struct {
	ID         string            `json:"id"`
	TenantID   string            `json:"tenantId"`
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	Quota      QuotaSpec         `json:"quota"`
	LimitRange LimitRangeSpec    `json:"limitRange"`
	CreatedAt  time.Time         `json:"createdAt"`
}

// TenantServiceAccount is a Kubernetes service account managed for a tenant namespace.
// It links our service account to tenant namespace
type TenantServiceAccount struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	NamespaceID string    `json:"namespaceId"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
}

// QuotaSpec is our simplified product model for a Kubernetes ResourceQuota.
// Values are strings because Kubernetes quantities use formats like "500m",
// "4", "8Gi", and "100Gi".
type QuotaSpec struct {
	RequestsCPU    string `json:"requestsCpu"`
	RequestsMemory string `json:"requestsMemory"`
	LimitsCPU      string `json:"limitsCpu"`
	LimitsMemory   string `json:"limitsMemory"`
	Pods           string `json:"pods"`
	PVCs           string `json:"pvcs"`
	Storage        string `json:"storage"`
}

// LimitRangeSpec is our simplified product model for a Kubernetes LimitRange.
// It defines default requests/limits and max per-container values inside one
// namespace.
type LimitRangeSpec struct {
	DefaultCPU    string `json:"defaultCpu"`
	DefaultMemory string `json:"defaultMemory"`
	RequestCPU    string `json:"requestCpu"`
	RequestMemory string `json:"requestMemory"`
	MaxCPU        string `json:"maxCpu"`
	MaxMemory     string `json:"maxMemory"`
}

// RoleTemplate is our UI/domain representation of RBAC permissions.
// It later becomes a Kubernetes Role or ClusterRole, depending on Scope.
type RoleTemplate struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenantId"`
	Name      string     `json:"name"`
	Scope     string     `json:"scope"`
	Rules     []RoleRule `json:"rules"`
	CreatedAt time.Time  `json:"createdAt"`
}

// RoleRule is one RBAC rule in our domain model.
// It maps directly to Kubernetes rbacv1.PolicyRule: apiGroups, resources, verbs.
type RoleRule struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// Assignment means "bind this subject to this role in this namespace".
// It later becomes a Kubernetes RoleBinding. SubjectKind is usually User,
// Group, or ServiceAccount.
type Assignment struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	NamespaceID string    `json:"namespaceId"`
	RoleID      string    `json:"roleId"`
	SubjectKind string    `json:"subjectKind"`
	SubjectName string    `json:"subjectName"`
	CreatedAt   time.Time `json:"createdAt"`
}

// KubeconfigIssue represents a request to create access credentials.
// The token itself is returned once and is not stored in local state.
type KubeconfigIssue struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenantId"`
	NamespaceID string     `json:"namespaceId"`
	Name        string     `json:"name"`
	TTLHours    int        `json:"ttlHours"`
	Kubeconfig  string     `json:"kubeconfig"`
	CreatedAt   time.Time  `json:"createdAt"`
	IssuedAt    *time.Time `json:"issuedAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

type KubeconfigIssueResult struct {
	Issue KubeconfigIssue `json:"issue"`
	YAML  string          `json:"yaml"`
}

// AuditEvent is an append-only record of a user-visible action.
// It is for traceability: who/what created or changed tenant resources.
type AuditEvent struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"createdAt"`
}
