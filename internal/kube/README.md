# internal/kube

Kubernetes integration package.

This package converts domain objects into typed Kubernetes resources and talks to the Kubernetes API through `client-go`.

## Files

- `client.go` - in-cluster/out-of-cluster client creation, cluster metadata, TokenRequest kubeconfig rendering.
- `builder.go` - typed Kubernetes object builders and YAML serialization.
- `apply.go` - create managed resources, ignoring already-existing objects.
- `delete.go` - delete managed resources with label protection.
- `get.go` - list managed cluster objects for observability.
- `types.go` - cluster snapshot DTOs.
- `utils.go` - shared Kubernetes API helpers.

## Rule

Prefer typed Kubernetes structs over hand-written YAML strings. YAML is an output format, not the internal representation.
