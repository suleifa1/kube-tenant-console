package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeClient struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

type ClusterSnapshot struct {
	Objects []ClusterObject `json:"objects"`
	Summary ClusterSummary  `json:"summary"`
}

type ClusterSummary struct {
	Namespaces      int `json:"namespaces"`
	ResourceQuotas  int `json:"resourceQuotas"`
	LimitRanges     int `json:"limitRanges"`
	ServiceAccounts int `json:"serviceAccounts"`
	Roles           int `json:"roles"`
	RoleBindings    int `json:"roleBindings"`
}

type ClusterObject struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId,omitempty"`
	Status    string `json:"status,omitempty"`
}
