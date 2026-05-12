package kube

import (
	"context"
	"fmt"

	"kube-tenant-console/internal/domain"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func NewKubeClientAuto(kubeconfig string) (*KubeClient, error) {
	if kubeconfig != "" {
		return NewKubeClientOutOfCluster(kubeconfig)
	}

	client, err := NewKubeClientInCluster()
	if err == nil {
		return client, nil
	}
	return NewKubeClientOutOfCluster("")
}

func NewKubeClientOutOfCluster(kubeconfig string) (*KubeClient, error) {
	if kubeconfig == "" {
		home := homedir.HomeDir()
		if home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubeClient{Clientset: clientset, Config: config}, nil
}

func NewKubeClientInCluster() (*KubeClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &KubeClient{Clientset: clientset, Config: config}, nil
}

func (client *KubeClient) IssueKubeconfigToken(ctx context.Context, namespace string, serviceAccountName string, tokenRequest *authv1.TokenRequest) (*authv1.TokenRequest, error) {
	return client.Clientset.CoreV1().ServiceAccounts(namespace).CreateToken(ctx, serviceAccountName, tokenRequest, metav1.CreateOptions{})
}

func (client *KubeClient) RenderKubeconfig(issue domain.KubeconfigIssue, ns domain.TenantNamespace, serviceAccount domain.TenantServiceAccount, tenant domain.Tenant, token string) (string, error) {
	serverURL := client.ClusterServer()
	if serverURL == "" {
		return "", fmt.Errorf("cluster server is empty")
	}
	caData, err := client.ClusterCAData()
	if err != nil {
		return "", err
	}
	cfg := BuildKubeConfigNamespace(ns, serviceAccount, issue, tenant, serverURL, caData, token)
	return KubeConfigToYAML(cfg)
}

func (client *KubeClient) ClusterServer() string {
	if client == nil || client.Config == nil {
		return ""
	}
	return client.Config.Host
}

func (client *KubeClient) ClusterCAData() ([]byte, error) {
	if client == nil || client.Config == nil {
		return nil, nil
	}
	if len(client.Config.CAData) > 0 {
		return client.Config.CAData, nil
	}
	if client.Config.CAFile != "" {
		return os.ReadFile(client.Config.CAFile)
	}
	return nil, nil
}
