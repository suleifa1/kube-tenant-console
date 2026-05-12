package kube

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubeClientOutOfClusterEnsureNamespace(t *testing.T) {
	if os.Getenv("LRBAC_KUBE_INTEGRATION") != "1" {
		t.Skip("set LRBAC_KUBE_INTEGRATION=1 to run against the current kubeconfig context")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := NewKubeClientOutOfCluster(os.Getenv("KUBECONFIG"))
	if err != nil {
		t.Fatalf("new out-of-cluster kube client: %v", err)
	}

	suffix := strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
	tenantID := "tenant-it-" + suffix
	namespaceName := "lrbac-it-" + suffix
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				ManagedByLabel: AppName,
				PartOfLabel:    AppName,
				TenantIDLabel:  tenantID,
			},
		},
	}

	defer func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := client.Clientset.CoreV1().Namespaces().Delete(deleteCtx, namespaceName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Logf("cleanup namespace %s: %v", namespaceName, err)
		}
	}()

	if err := client.EnsureNamespace(ctx, namespace); err != nil {
		t.Fatalf("ensure namespace first call: %v", err)
	}
	if err := client.EnsureNamespace(ctx, namespace); err != nil {
		t.Fatalf("ensure namespace second call should tolerate existing namespace: %v", err)
	}

	list, err := client.GetNamespaces(ctx, map[string]string{
		ManagedByLabel: AppName,
		TenantIDLabel:  tenantID,
	})
	if err != nil {
		t.Fatalf("list managed namespaces: %v", err)
	}

	for _, item := range list.Items {
		if item.Name == namespaceName {
			return
		}
	}
	t.Fatalf("created namespace %q was not returned by managed label selector", namespaceName)
}
