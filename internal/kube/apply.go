package kube

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-tenant-console/internal/domain"
)

func (client *KubeClient) EnsureNamespace(ctx context.Context, namespace *corev1.Namespace) error {
	_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	return ignoreAlreadyExists(err)
}

func (client *KubeClient) EnsureNamespaceBundle(ctx context.Context, ns domain.TenantNamespace) error {
	if err := client.EnsureNamespace(ctx, BuildNamespace(ns)); err != nil {
		return err
	}
	if err := client.EnsureResourceQuota(ctx, BuildResourceQuota(ns)); err != nil {
		return err
	}
	return client.EnsureLimitRange(ctx, BuildLimitRange(ns))
}

func (client *KubeClient) EnsureResourceQuota(ctx context.Context, quota *corev1.ResourceQuota) error {
	_, err := client.Clientset.CoreV1().ResourceQuotas(quota.Namespace).Create(ctx, quota, metav1.CreateOptions{})
	return ignoreAlreadyExists(err)
}

func (client *KubeClient) EnsureLimitRange(ctx context.Context, limitRange *corev1.LimitRange) error {
	_, err := client.Clientset.CoreV1().LimitRanges(limitRange.Namespace).Create(ctx, limitRange, metav1.CreateOptions{})
	return ignoreAlreadyExists(err)
}

func (client *KubeClient) EnsureServiceAccount(ctx context.Context, serviceAccount *corev1.ServiceAccount) error {
	_, err := client.Clientset.CoreV1().ServiceAccounts(serviceAccount.Namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	return ignoreAlreadyExists(err)
}

func (client *KubeClient) EnsureRole(ctx context.Context, role *rbacv1.Role) error {
	_, err := client.Clientset.RbacV1().Roles(role.Namespace).Create(ctx, role, metav1.CreateOptions{})
	return ignoreAlreadyExists(err)
}

func (client *KubeClient) EnsureRoleBinding(ctx context.Context, roleBinding *rbacv1.RoleBinding) error {
	_, err := client.Clientset.RbacV1().RoleBindings(roleBinding.Namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	return ignoreAlreadyExists(err)
}
