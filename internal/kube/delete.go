package kube

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *KubeClient) DeleteNamespace(ctx context.Context, name string, selectors map[string]string) error {
	namespace, err := client.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !hasLabels(namespace.Labels, selectors) {
		return fmt.Errorf("namespace %s does not match required managed labels", name)
	}

	var gracePeriodSeconds int64 = 90
	return client.Clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
}

func (client *KubeClient) DeleteServiceAccount(ctx context.Context, namespace, name string, selectors map[string]string) error {
	object, err := client.Clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !hasLabels(object.Labels, selectors) {
		return fmt.Errorf("service account %s/%s does not match required managed labels", namespace, name)
	}
	return client.Clientset.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (client *KubeClient) DeleteRole(ctx context.Context, namespace, name string, selectors map[string]string) error {
	object, err := client.Clientset.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !hasLabels(object.Labels, selectors) {
		return fmt.Errorf("role %s/%s does not match required managed labels", namespace, name)
	}
	return client.Clientset.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (client *KubeClient) DeleteRoleBinding(ctx context.Context, namespace, name string, selectors map[string]string) error {
	object, err := client.Clientset.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !hasLabels(object.Labels, selectors) {
		return fmt.Errorf("role binding %s/%s does not match required managed labels", namespace, name)
	}
	return client.Clientset.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (client *KubeClient) DeleteResourceQuota(ctx context.Context, namespace, name string, selectors map[string]string) error {
	object, err := client.Clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !hasLabels(object.Labels, selectors) {
		return fmt.Errorf("resource quota %s/%s does not match required managed labels", namespace, name)
	}
	return client.Clientset.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (client *KubeClient) DeleteLimitRange(ctx context.Context, namespace, name string, selectors map[string]string) error {
	object, err := client.Clientset.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !hasLabels(object.Labels, selectors) {
		return fmt.Errorf("limit range %s/%s does not match required managed labels", namespace, name)
	}
	return client.Clientset.CoreV1().LimitRanges(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
