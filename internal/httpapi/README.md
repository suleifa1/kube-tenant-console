# internal/httpapi

HTTP API and embedded UI package.

This package maps HTTP requests to local state operations and Kubernetes operations. It should stay thin: validation and mutations belong to `domain` and `local`, Kubernetes object logic belongs to `kube`.

## Files

- `server.go` - server wiring, shared response helpers, kube client guard.
- `server_get.go` - state, cluster snapshot, YAML preview endpoints.
- `server_post.go` - create/apply/token endpoints.
- `server_delete.go` - local delete and cluster delete endpoints.
- `types.go` - local aliases for cross-package types used by handlers.
- `static.go` - embedded static file serving.
- `static/` - browser UI.

## UI Note

The UI is a generated/prototyped admin interface. It is useful for operating the tool, but the durable project boundary is the HTTP API plus the Go packages.
