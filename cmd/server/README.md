# cmd/server

Process entrypoint.

It reads environment configuration, opens the local JSON store, tries to create a Kubernetes client, wires the HTTP server, and starts `net/http`.

## Inputs

- `ADDR` - listen address, default `:8080`.
- `DATA_PATH` - JSON state path, default `./data/state.json`.
- `KUBECONFIG` - optional kubeconfig path for out-of-cluster mode.
- `ALLOW_CLUSTER_SCOPE` - enables cluster-scoped role templates when set to `true`.

## Notes

The server can start even when the Kubernetes client is unavailable. In that mode local state and YAML preview still work, but cluster apply/delete/token operations return an API error.
