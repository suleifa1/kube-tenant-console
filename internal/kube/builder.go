package kube

import (
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	cmdapi "k8s.io/client-go/tools/clientcmd/api"
	"kube-tenant-console/internal/domain"
	"sigs.k8s.io/yaml"
)

func BuildNamespace(ns domain.TenantNamespace) *corev1.Namespace {
	labels := tenantLabels(ns)

	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns.Name,
			Labels: labels,
		},
	}
}

func BuildResourceQuota(ns domain.TenantNamespace) *corev1.ResourceQuota {
	labels := tenantLabels(ns)
	return &corev1.ResourceQuota{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ResourceQuota",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-quota",
			Namespace: ns.Name,
			Labels:    labels,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsCPU:            res.MustParse(ns.Quota.RequestsCPU),
				corev1.ResourceRequestsMemory:         res.MustParse(ns.Quota.RequestsMemory),
				corev1.ResourceLimitsCPU:              res.MustParse(ns.Quota.LimitsCPU),
				corev1.ResourceLimitsMemory:           res.MustParse(ns.Quota.LimitsMemory),
				corev1.ResourcePods:                   res.MustParse(ns.Quota.Pods),
				corev1.ResourcePersistentVolumeClaims: res.MustParse(ns.Quota.PVCs),
				corev1.ResourceRequestsStorage:        res.MustParse(ns.Quota.Storage),
			},
		},
	}
}

func BuildLimitRange(ns domain.TenantNamespace) *corev1.LimitRange {
	labels := tenantLabels(ns)
	return &corev1.LimitRange{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "LimitRange",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-defaults",
			Namespace: ns.Name,
			Labels:    labels,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type: corev1.LimitTypeContainer,
					Default: corev1.ResourceList{
						corev1.ResourceCPU:    res.MustParse(ns.LimitRange.DefaultCPU),
						corev1.ResourceMemory: res.MustParse(ns.LimitRange.DefaultMemory),
					},
					DefaultRequest: corev1.ResourceList{
						corev1.ResourceCPU:    res.MustParse(ns.LimitRange.RequestCPU),
						corev1.ResourceMemory: res.MustParse(ns.LimitRange.RequestMemory),
					},
					Max: corev1.ResourceList{
						corev1.ResourceCPU:    res.MustParse(ns.LimitRange.MaxCPU),
						corev1.ResourceMemory: res.MustParse(ns.LimitRange.MaxMemory),
					},
				},
			},
		},
	}
}

func BuildRole(role domain.RoleTemplate, ns domain.TenantNamespace) *rbacv1.Role {
	labels := tenantLabels(ns)

	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      role.Name,
			Namespace: ns.Name,
			Labels:    labels,
		},
		Rules: buildPolicyRules(role.Rules),
	}
}

func BuildRoleBind(assignment domain.Assignment, role domain.RoleTemplate, ns domain.TenantNamespace) *rbacv1.RoleBinding {
	kind := "RoleBinding"
	roleKind := "Role"
	labels := tenantLabels(ns)
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      RoleBindingName(assignment, role),
			Namespace: ns.Name,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			buildSubject(assignment, ns),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     role.Name,
		},
	}
}

func RoleBindingName(assignment domain.Assignment, role domain.RoleTemplate) string {
	if assignment.ID == "" {
		return role.Name + "-binding"
	}
	return role.Name + "-" + assignment.ID
}

func BuildServiceAccount(ns domain.TenantNamespace, tsa domain.TenantServiceAccount) *corev1.ServiceAccount {
	labels := tenantLabels(ns)
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tsa.Name,
			Namespace: ns.Name,
			Labels:    labels,
		},
	}
}

func BuildKubeConfigNamespace(ns domain.TenantNamespace, sa domain.TenantServiceAccount, issue domain.KubeconfigIssue, tn domain.Tenant, serverURL string, caBytes []byte, token string) cmdapi.Config {
	clusterName := tn.Name + "-cluster"
	if tn.Name == "" {
		clusterName = AppName + "-cluster"
	}
	userName := issue.Name
	if userName == "" {
		userName = sa.Name
	}
	if token == "" {
		token = "TOKEN_WILL_BE_ISSUED_BY_TOKENREQUEST"
	}

	cfg := cmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*cmdapi.Cluster{
			clusterName: {
				Server:                   serverURL,
				CertificateAuthorityData: caBytes,
			},
		},
		AuthInfos: map[string]*cmdapi.AuthInfo{
			userName: {
				Token: token,
			},
		},
		Contexts: map[string]*cmdapi.Context{
			ns.Name: {
				Cluster:   clusterName,
				AuthInfo:  userName,
				Namespace: ns.Name,
			},
		},
		CurrentContext: ns.Name,
	}
	return cfg
}

func BuildServiceAccountTokenRequest(ttlHours int) *authv1.TokenRequest {
	seconds := int64(ttlHours) * 60 * 60
	return &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			ExpirationSeconds: &seconds,
		},
	}
}

func KubeConfigToYAML(cfg cmdapi.Config) (string, error) {
	raw, err := clientcmd.Write(cfg)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func buildSubject(assignment domain.Assignment, ns domain.TenantNamespace) rbacv1.Subject {
	subject := rbacv1.Subject{
		Kind: assignment.SubjectKind,
		Name: assignment.SubjectName,
	}

	if assignment.SubjectKind == "ServiceAccount" {
		subject.Namespace = ns.Name
		return subject
	}

	subject.APIGroup = "rbac.authorization.k8s.io"
	return subject
}

func buildPolicyRules(rules []domain.RoleRule) []rbacv1.PolicyRule {
	out := make([]rbacv1.PolicyRule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, rbacv1.PolicyRule{
			APIGroups: rule.APIGroups,
			Resources: rule.Resources,
			Verbs:     rule.Verbs,
		})
	}
	return out
}

func tenantLabels(ns domain.TenantNamespace) map[string]string {
	return map[string]string{
		ManagedByLabel: AppName,
		PartOfLabel:    AppName,
		TenantIDLabel:  ns.TenantID,
	}
}

func ObjectToYAML(obj runtime.Object) string {
	raw, err := yaml.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(raw)
}
