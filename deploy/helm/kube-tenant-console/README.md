# Helm Chart

Deploys Kube Tenant Console into a Kubernetes cluster.

## What It Creates

- `Deployment`
- `Service`
- `ServiceAccount`
- optional privileged RBAC binding
- optional `PersistentVolumeClaim` for `state.json`
- `ConfigMap` with runtime environment values

## Default Access

The chart exposes a `ClusterIP` service. Use port-forward for local admin access:

```sh
kubectl port-forward -n kube-tenant-console svc/kube-tenant-console 8080:80
```

## Values

- `image.*` - container image settings.
- `serviceAccount.clusterAdmin` - grants the app broad cluster permissions when `true`.
- `env.addr` - process listen address inside the pod.
- `env.dataPath` - JSON state file path.
- `env.allowClusterScope` - enables cluster-scoped role templates.
- `persistence.*` - PVC settings for the local state file.

The chart is designed for one replica when using the PVC JSON store.
