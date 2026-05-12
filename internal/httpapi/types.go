package httpapi

import (
	"kube-tenant-console/internal/domain"
	kubemod "kube-tenant-console/internal/kube"
	"kube-tenant-console/internal/local"
)

type Store = local.Store
type Config = domain.Config
type KubeClient = kubemod.KubeClient
type State = domain.State
type Tenant = domain.Tenant
type TenantNamespace = domain.TenantNamespace
type TenantServiceAccount = domain.TenantServiceAccount
type QuotaSpec = domain.QuotaSpec
type LimitRangeSpec = domain.LimitRangeSpec
type RoleTemplate = domain.RoleTemplate
type RoleRule = domain.RoleRule
type Assignment = domain.Assignment
type KubeconfigIssue = domain.KubeconfigIssue
type KubeconfigIssueResult = domain.KubeconfigIssueResult

type Server struct {
	store *Store
	cfg   Config
	kube  *KubeClient
}

var BuildNamespace = kubemod.BuildNamespace
var BuildResourceQuota = kubemod.BuildResourceQuota
var BuildLimitRange = kubemod.BuildLimitRange
var BuildRole = kubemod.BuildRole
var BuildRoleBind = kubemod.BuildRoleBind
var RoleBindingName = kubemod.RoleBindingName
var BuildServiceAccount = kubemod.BuildServiceAccount
var BuildServiceAccountTokenRequest = kubemod.BuildServiceAccountTokenRequest
var ObjectToYAML = kubemod.ObjectToYAML
var ManagedSelectors = kubemod.ManagedSelectors
