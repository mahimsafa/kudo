# CLI Usage Reference

The `kudo` binary has two roles:

- **`kudo agent`** — long-running process on each server (cluster member, workloads, L7 proxy).
- **All other commands** — clients that talk to the agent gRPC API (default `127.0.0.1:9090`).

Run `kudo --help` or `kudo <command> --help` for the latest flag list.

## Global options

| Flag / env | Description |
|------------|-------------|
| `--server <host:port>` | gRPC API address (default `127.0.0.1:9090`) |
| `KUDO_SERVER` | Same as `--server` when the flag is not set |

Client commands: `apply`, `remove`, `status`, `scale`, `nodes`. They fail with *connection refused* if no agent is listening on the API port.

---

## `kudo agent`

Starts the Kudo agent on this machine.

```bash
kudo agent --bootstrap --name node-1
kudo agent --join 10.0.0.1:7946 --token <token> --name node-2
kudo agent -c /etc/kudo/kudo.yaml --bootstrap --name node-1
```

| Flag | Description |
|------|-------------|
| `-c, --config` | Agent config file (default `/etc/kudo/kudo.yaml`). If missing, local dev defaults apply (`~/.kudo/data`, proxy port `8088`). |
| `--bootstrap` | Create a new single-node (or first) cluster |
| `--join` | Gossip addresses of existing nodes (repeatable) |
| `--token` | Join token from `kudo token create` |
| `--name` | Node name (default: hostname) |

**Notes**

- Leave the process running in a terminal or under systemd (see [Deploy a Web Application](deploy-web-application.md)).
- Only the Raft **leader** applies `apply` / `remove` / `scale` changes to cluster state.
- Logs go to stderr (structured JSON with zap). Use `log.level: debug` in agent config for more detail.

---

## `kudo apply`

Create or update applications from a YAML manifest.

```bash
kudo apply -f configs/examples/docker-app.yaml
kudo apply -f ./my-app.yaml --server 10.0.0.1:9090
```

| Flag | Description |
|------|-------------|
| `-f, --file` | Path to YAML file (required). Supports multiple documents separated by `---`. |

**Behavior**

1. Parses each `Application` document.
2. Writes desired state to the cluster (Raft) when this node is leader.
3. The **reconciler** (about every 10s) starts/stops workloads to match `replicas`.
4. The **L7 proxy** routes traffic when `routing.domain` is set (see [Application Manifest](application-manifest.md)).

**Typical failures**

| Message | What to do |
|---------|------------|
| `connection refused` | Start `kudo agent` first; wait a few seconds for leader election |
| `not cluster leader` | Retry `apply` on the leader node or wait for election |
| `parsing config` | Fix YAML; every app needs `name`, `adapter`, and valid `spec` |

After success, run `kudo status <app-name>` and wait ~10–15s for replicas to become running.

---

## `kudo remove`

Remove applications listed in a manifest file.

```bash
kudo remove -f configs/examples/docker-app.yaml
kudo remove -f my-app.yaml -y
```

| Flag | Description |
|------|-------------|
| `-f, --file` | Same shape as `apply` (required) |
| `-y, --yes` | If containers are already gone, skip confirmation and remove cluster state |

**Behavior**

1. Checks **dependencies** (another app still using the same `routing.domain` + `routing.path` blocks removal).
2. Stops local Docker containers for each instance.
3. Deletes application and instances from cluster state; removes proxy routes.

If a container was removed manually, you get **warnings** and a prompt:

```text
Remove cluster state, routes, and any remaining resources anyway? [y/N]:
```

Use `-y` in scripts. **Nothing is changed** until you confirm (or pass `-y`).

---

## `kudo status`

Show application status.

```bash
kudo status
kudo status nginx-demo
```

| Form | Output |
|------|--------|
| No args | Table of all applications (name, adapter, desired replicas) |
| `<app-name>` | Desired vs running replicas and instance list (ID, node, status, address) |

**Instance `address`** — backend URL for the proxy (e.g. `127.0.0.1:49153`), not necessarily port 80 on the host.

If `0/N running`, check Docker and agent logs (`action failed`, `pulling image`, etc.).

---

## `kudo scale`

Change replica count without editing the YAML file.

```bash
kudo scale nginx-demo --replicas 3
kudo scale nginx-demo --replicas 0
```

| Flag | Description |
|------|-------------|
| `--replicas` | Target count (required) |

Reconciliation runs on the next loop (~10s). Scaling to `0` stops workloads but keeps the app in `kudo status` until you `remove` it.

---

## `kudo nodes`

List nodes registered in cluster state.

```bash
kudo nodes
```

Shows ID, name, address, and status (e.g. `healthy`). The local node registers itself when it becomes Raft leader.

---

## `kudo token create`

Generate a join token (run on a machine with the agent running).

```bash
kudo token create --ttl 24h
kudo token create --ttl 48h
```

| Flag | Description |
|------|-------------|
| `--ttl` | Token lifetime (default `24h`) |

Use the printed token with `kudo agent --join ... --token <token>` on new nodes.

---

## Common workflows

### Local development (two terminals)

```bash
# Terminal 1
make build
./bin/kudo agent --bootstrap --name dev-node-1

# Terminal 2
./bin/kudo apply -f configs/examples/docker-app.yaml
./bin/kudo status nginx-demo
curl http://127.0.0.1:8088/    # local dev proxy port; see application manifest
```

### Production-style (single leader, remote CLI)

```bash
export KUDO_SERVER=10.0.0.1:9090
kudo apply -f prod/my-app.yaml
kudo status my-app
```

See [Deploy a Web Application](deploy-web-application.md) for the full production checklist.

---

## Related docs

- [Application Manifest](application-manifest.md) — YAML fields for deployments
- [Agent Configuration](configuration.md) — `kudo.yaml` for each server
- [Deploy a Web Application](deploy-web-application.md) — end-to-end production guide
