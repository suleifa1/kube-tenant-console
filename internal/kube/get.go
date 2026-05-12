package kube

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *KubeClient) GetNamespaces(ctx context.Context, selectors map[string]string) (*corev1.NamespaceList, error) {
	return client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector(selectors),
	})
}

// GetManagedClusterSnapshot returns the live Kubernetes objects managed by lrbac labels.
func (client *KubeClient) GetManagedClusterSnapshot(ctx context.Context) (ClusterSnapshot, error) {
	var snapshot ClusterSnapshot
	selectors := ManagedSelectors()
	selector := labelSelector(selectors)

	namespaces, err := client.GetNamespaces(ctx, selectors)
	if err != nil {
		return snapshot, err
	}

	for _, namespace := range namespaces.Items {
		snapshot.Summary.Namespaces++
		snapshot.Objects = append(snapshot.Objects, ClusterObject{
			Kind:     "Namespace",
			Name:     namespace.Name,
			TenantID: namespace.Labels[TenantIDLabel],
			Status:   string(namespace.Status.Phase),
		})

		quotas, err := client.Clientset.CoreV1().ResourceQuotas(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return snapshot, err
		}
		for _, quota := range quotas.Items {
			snapshot.Summary.ResourceQuotas++
			snapshot.Objects = append(snapshot.Objects, clusterObject("ResourceQuota", namespace.Name, quota.Name, quota.Labels))
		}

		limitRanges, err := client.Clientset.CoreV1().LimitRanges(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return snapshot, err
		}
		for _, limitRange := range limitRanges.Items {
			snapshot.Summary.LimitRanges++
			snapshot.Objects = append(snapshot.Objects, clusterObject("LimitRange", namespace.Name, limitRange.Name, limitRange.Labels))
		}

		serviceAccounts, err := client.Clientset.CoreV1().ServiceAccounts(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return snapshot, err
		}
		for _, serviceAccount := range serviceAccounts.Items {
			snapshot.Summary.ServiceAccounts++
			snapshot.Objects = append(snapshot.Objects, clusterObject("ServiceAccount", namespace.Name, serviceAccount.Name, serviceAccount.Labels))
		}

		roles, err := client.Clientset.RbacV1().Roles(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return snapshot, err
		}
		for _, role := range roles.Items {
			snapshot.Summary.Roles++
			snapshot.Objects = append(snapshot.Objects, clusterObject("Role", namespace.Name, role.Name, role.Labels))
		}

		roleBindings, err := client.Clientset.RbacV1().RoleBindings(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return snapshot, err
		}
		for _, roleBinding := range roleBindings.Items {
			snapshot.Summary.RoleBindings++
			snapshot.Objects = append(snapshot.Objects, clusterObject("RoleBinding", namespace.Name, roleBinding.Name, roleBinding.Labels))
		}
	}

	return snapshot, nil
}

func (client *KubeClient) GetServiceAccounts(ctx context.Context, selectors map[string]string, nsList *corev1.NamespaceList) (map[string]*corev1.ServiceAccountList, error) {
	if nsList == nil {
		return nil, fmt.Errorf("namespaces are empty")
	}

	all := make(map[string]*corev1.ServiceAccountList, len(nsList.Items))
	selector := labelSelector(selectors)
	for _, namespace := range nsList.Items {
		serviceAccounts, err := client.Clientset.CoreV1().ServiceAccounts(namespace.Name).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		all[namespace.Name] = serviceAccounts
	}
	return all, nil
}

func (client *KubeClient) GetRoles(ctx context.Context, selectors map[string]string, nsList *corev1.NamespaceList) (map[string]*rbacv1.RoleList, error) {
	if nsList == nil {
		return nil, fmt.Errorf("namespaces are empty")
	}

	all := make(map[string]*rbacv1.RoleList, len(nsList.Items))
	selector := labelSelector(selectors)
	for _, namespace := range nsList.Items {
		roles, err := client.Clientset.RbacV1().Roles(namespace.Name).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		all[namespace.Name] = roles
	}
	return all, nil
}

func (client *KubeClient) GetRoleBindings(ctx context.Context, selectors map[string]string, nsList *corev1.NamespaceList) (map[string]*rbacv1.RoleBindingList, error) {
	if nsList == nil {
		return nil, fmt.Errorf("namespaces are empty")
	}

	all := make(map[string]*rbacv1.RoleBindingList, len(nsList.Items))
	selector := labelSelector(selectors)
	for _, namespace := range nsList.Items {
		roleBindings, err := client.Clientset.RbacV1().RoleBindings(namespace.Name).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		all[namespace.Name] = roleBindings
	}
	return all, nil
}
