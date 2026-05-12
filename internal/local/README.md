# internal/local

Local state package.

This package owns the JSON-backed store and all mutations of the app state. HTTP handlers call this package instead of editing `domain.State` directly.

## Files

- `store.go` - open, snapshot, update, normalize, save.
- `helpers.go` - IDs, audit records, find helpers, kubeconfig issue resolution.
- `mutations.go` - create tenant, namespace, role, service account, assignment, kubeconfig request.
- `deletions.go` - remove local objects and dependent state.

## Storage

State is stored as one JSON document. The store uses a process-local mutex around updates, so the expected deployment is one app replica per state file.
