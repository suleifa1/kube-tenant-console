# internal/domain

Product model package.

This package defines the local state shape and validation rules used by the rest of the app. It does not talk to HTTP, disk, or the Kubernetes API.

## Files

- `models.go` - tenants, namespaces, roles, service accounts, bindings, kubeconfig requests, audit events.
- `defaults.go` - default quota and limit range values.
- `validation.go` - DNS label checks, subject kind checks, RBAC guardrails.

## Rule

Keep Kubernetes client objects out of this package. Domain structs are the app's own JSON/API model, not Kubernetes API structs.
