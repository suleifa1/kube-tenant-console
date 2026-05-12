package kube

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	AppName        = "lrbac"
	ManagedByLabel = "app.kubernetes.io/managed-by"
	PartOfLabel    = "app.kubernetes.io/part-of"
	TenantIDLabel  = "lrbac.dev/tenant-id"
	TenantLabel    = "lrbac.dev/tenant"
)

func ManagedSelectors() map[string]string {
	return map[string]string{
		ManagedByLabel: AppName,
		PartOfLabel:    AppName,
	}
}

func labelSelector(selectors map[string]string) string {
	return labels.Set(selectors).AsSelector().String()
}

func ignoreAlreadyExists(err error) error {
	if err == nil || apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func clusterObject(kind, namespace, name string, objectLabels map[string]string) ClusterObject {
	return ClusterObject{
		Kind:      kind,
		Namespace: namespace,
		Name:      name,
		TenantID:  objectLabels[TenantIDLabel],
	}
}

func hasLabels(objectLabels map[string]string, selectors map[string]string) bool {
	for key, value := range selectors {
		if objectLabels[key] != value {
			return false
		}
	}
	return true
}
