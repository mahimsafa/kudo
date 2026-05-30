# Deploy a Web Application (End-to-End)

This guide walks through deploying a stateless web app (for example nginx or a containerized API) from first cluster setup through production access, including **logging**, **retries**, **failure handling**, and **gaps** you should plan around today.

**Prerequisites:** Linux or macOS servers, Docker on workload nodes, network between nodes, DNS for public hostnames.

**Related docs:** [CLI Usage](cli-usage.md) · [Application Manifest](application-manifest.md) · [Agent Configuration](configuration.md)

---

## Overview

| Layer | Responsibility |
|-------|----------------|
| **CLI** (`apply`, `status`, `remove`, …) | Sends desired state to the leader agent |
| **Raft / FSM** | Stores applications, nodes, instances |
| **Reconciler** | Every ~10s, scale workloads to match `replicas` |
| **Docker executor** | Pull image, run containers, publish ephemeral host ports |
| **L7 proxy** | Listens on `proxy.http_port`; round-robin to instance addresses |

You do **not** get direct “each replica on host port 80” by default. Users hit the **Kudo proxy port**, which forwards to container backends.

---

## Phase 1 — Install and bootstrap

### 1.1 Install Kudo on each server

```bash
curl -fsSL https://raw.githubusercontent.com/mahimsafa/kudo/main/install.sh | bash -s -- --system --version v0.1.0
# or: make build && sudo cp bin/kudo /usr/local/bin/
```

Install **Docker** on every node that should run containers.

### 1.2 Agent config (production)

On each server, create `/etc/kudo/kudo.yaml` (see [Agent Configuration](configuration.md)):

```yaml
node:
  name: "node-1"
  bind_addr: "0.0.0.0"
  bind_port: 7946
  advertise_addr: "<this-server-lan-ip>"
  data_dir: "/var/lib/kudo"

cluster:
  bootstrap: true   # first node only
  join_addrs: []    # joiners: ["10.0.0.1:7946"]

api:
  grpc_port: 9090

proxy:
  http_port: 80
  https_port: 443

log:
  level: "info"     # use "debug" while troubleshooting
```

**First node:**

```bash
sudo kudo agent --bootstrap --name node-1 -c /etc/kudo/kudo.yaml
```

**Additional nodes:**

```bash
# On leader: generate token
kudo token create --ttl 24h

# On joiner:
sudo kudo agent --join 10.0.0.1:7946 --token <token> --name node-2 -c /etc/kudo/kudo.yaml
```

Verify:

```bash
kudo nodes
```

### 1.3 Run agent under process supervision (recommended)

Run the agent under **systemd** (or another supervisor) so it restarts on failure. The install script may place unit files depending on options—confirm for your install path.

**Fallback if agent exits:** Restart the service; Raft state is under `data_dir`. Workloads may keep running in Docker until reconciler runs again.

---

## Phase 2 — Write the application manifest

Example production-style nginx:

```yaml
kind: Application
name: nginx-demo
adapter: docker
replicas: 2
spec:
  image: nginx:alpine
  ports:
    - port: 80
routing:
  domain: www.example.com
  path: /
  ingress_port: 80
  algorithm: round-robin
```

For an app listening on **8080** exposed as **80** publicly, see [`configs/examples/docker-app-public-port.yaml`](../configs/examples/docker-app-public-port.yaml).

Field reference: [Application Manifest](application-manifest.md).

---

## Phase 3 — Deploy (`apply`)

From a machine that can reach the leader’s gRPC API:

```bash
export KUDO_SERVER=<leader-ip>:9090   # optional
kudo apply -f nginx-demo.yaml
```

**Expected output**

```text
Applied successfully: applied 1 application
Wait ~10s for replicas, then run: kudo status <app-name>
```

### 3.1 Leader election

If you see `not cluster leader`, wait 2–5 seconds and **retry `apply`**. Only the Raft leader accepts writes.

### 3.2 Wait for reconciliation (retry loop)

The reconciler runs about **every 10 seconds**. This is the built-in **retry** mechanism for failed deploys:

| Attempt | What happens |
|---------|----------------|
| T+0s | `apply` writes desired state |
| T+10s, T+20s, … | Reconciler deploys missing instances or stops excess ones |

**Check progress:**

```bash
watch -n 2 'kudo status nginx-demo'
```

Target: `Replicas: 2/2 running` with non-empty instance `ADDRESS` values.

**Fallback timeline**

1. `apply` succeeded but `0/N running` → wait at least 2 reconcile intervals (~20s).
2. Still zero → inspect logs (Phase 5); fix Docker/image/pull issues.
3. Partial `1/2` → one deploy failed; reconciler retries next tick unless error is permanent (bad image name).

There is **no exponential backoff** or per-attempt CLI retry flag today—only periodic reconciler passes.

---

## Phase 4 — Networking and verify

### 4.1 DNS

Point `www.example.com` A/AAAA records at your nodes’ **public** IPs (or a load balancer in front of them—see TODO).

### 4.2 Open firewall

Allow:

- **TCP `proxy.http_port`** (e.g. 80) — user traffic
- **TCP `api.grpc_port`** (9090) — CLI/admin (restrict to admin IPs in prod)
- **TCP 7946** — gossip between nodes (cluster LAN/security group)
- **Raft port** — `grpc_port + 1000` on advertised address (default 10090) between nodes

### 4.3 Smoke test

On a node with the proxy listening:

```bash
curl -v -H 'Host: www.example.com' http://127.0.0.1/
```

From the internet:

```bash
curl -v http://www.example.com/
```

**Local dev note:** Without `/etc/kudo/kudo.yaml`, proxy uses port **8088** and `local_access: true` allows `curl http://127.0.0.1:8088/`. See [Getting Started](getting-started.md).

### 4.4 Load spread (round-robin)

Repeated curls should hit different backends when `replicas > 1` and `routing.algorithm: round-robin` (default behavior in proxy).

---

## Phase 5 — Error logging and diagnosis

### 5.1 CLI errors

| Symptom | Likely cause | Action |
|---------|--------------|--------|
| `connection refused` on `:9090` | Agent not running or wrong `KUDO_SERVER` | Start agent; check port |
| `not cluster leader` | Election in progress | Retry `apply` |
| `apply failed: parsing config` | Invalid YAML | Fix manifest |
| `remove blocked by dependencies` | Shared ingress route | Remove all sharing apps in one file or remove blocker first |

### 5.2 Agent logs (primary operations log)

Watch the agent process stderr/journal:

```bash
journalctl -u kudo -f    # if using systemd
```

| Log / message | Meaning |
|---------------|---------|
| `registered adapter name=docker` | Docker available |
| `docker adapter unavailable` | Docker socket missing; no containers will run |
| `reconciler action type=deploy` | Scheduling new instance |
| `pulling image` | Pull started |
| `action failed` + `Cannot connect to the Docker daemon` | Start Docker |
| `action failed` + `pulling image` / `creating container` | Fix image name, registry auth, disk |
| `container started` + `address` | Success; note backend host port |
| `proxy route updated` + `backends` | Proxy has backends; ingress should work |
| `application reachable via L7 proxy` | Suggested URLs (when domain set) |

Set `log.level: debug` for more detail, then return to `info` in prod.

### 5.3 Application status

```bash
kudo status nginx-demo
```

| Field | Healthy sign |
|-------|----------------|
| `Replicas` | `N/N running` with N = desired |
| `Instances` / `ADDRESS` | `127.0.0.1:<port>` per instance |
| `STATUS` | `running` |

### 5.4 Docker (fallback inspection)

```bash
docker ps --filter label=kudo.app=nginx-demo
```

If containers exist but `status` shows 0 running, cluster state may be out of sync—check agent logs for Raft/leader issues.

### 5.5 Proxy without backends

`curl` returns **404** from Kudo proxy → no route or empty backends.

- Confirm `routing.domain` matches `Host` header (or enable `local_access` for localhost tests).
- Confirm `proxy route updated` shows `backends > 0`.
- Confirm `ingress_port` / agent `proxy.http_port` match how you curl.

---

## Phase 6 — Scale and update

### Scale replicas

```bash
kudo scale nginx-demo --replicas 4
```

Wait for reconciler (~10s per adjustment). No rolling surge limits—new instances start as the loop schedules them.

### Update image / config

Change YAML and re-apply:

```bash
kudo apply -f nginx-demo.yaml
```

**Today:** Desired state updates in Raft; reconciler does **not** automatically recreate all containers for image-only changes unless replica count changes or instances are removed. **Fallback:** `kudo remove -f nginx-demo.yaml` then `apply` again, or scale to 0 then back (see TODO: rolling updates).

---

## Phase 7 — Teardown

```bash
kudo remove -f nginx-demo.yaml
```

- Missing containers → warnings + confirmation (or `-y`).
- Shared route with another app → blocked until dependency removed.

Verify:

```bash
kudo status
docker ps --filter label=kudo.app=nginx-demo
```

---

## Production checklist

Use this before go-live:

- [ ] Agent runs on all nodes with persistent `data_dir`
- [ ] `advertise_addr` reachable between nodes
- [ ] Docker running where `adapter: docker` workloads schedule
- [ ] `proxy.http_port` matches DNS/LB (usually 80)
- [ ] Firewall rules applied
- [ ] `apply` run against **leader** API
- [ ] `kudo status` shows all replicas running
- [ ] `proxy route updated` in logs with correct backends
- [ ] External `curl` succeeds with correct `Host`
- [ ] Join tokens and gRPC **not** exposed to the public internet
- [ ] Backup / DR plan for `data_dir` (Raft state)

---

## Production gaps (TODO)

Features you may expect in production that are **missing or incomplete** in Kudo today. Track these before hardening go-live.

### Security and cluster

- [ ] **Configurable join secret** — `kudo token create` uses a placeholder secret in code; production needs a real cluster secret store/config.
- [ ] **gRPC TLS and authentication** — CLI → agent is insecure gRPC by default.
- [ ] **Join token validation on agent** — verify end-to-end enforcement story for your version.
- [ ] **Firewall documentation / automation** in install script for Raft/gossip ports.

### Ingress and TLS

- [ ] **TLS termination** (`routing.tls: auto|manual`) — not implemented on proxy.
- [ ] **HTTPS listener** (`proxy.https_port`) — not wired.
- [ ] **ACME / certificate renewal** — not implemented.
- [ ] **External load balancer integration** doc (health checks for LB → Kudo).

### Health, routing, and load balancing

- [ ] **Active health checks** — `routing.healthcheck` stored but not used to remove unhealthy backends.
- [ ] **`least-connections` algorithm** — only round-robin implemented.
- [ ] **Path-based routing** beyond simple host+path key matching.
- [ ] **WebSocket / long-lived connection** tuning on reverse proxy.

### Workloads and releases

- [ ] **`nodejs` / `python` adapters** — not registered; Docker-only for real deploys.
- [ ] **`spec.entrypoint` / `spec.directory`** for Docker — not applied to container create.
- [ ] **Rolling updates** — changing image via `apply` does not roll pods automatically.
- [ ] **Registry credentials** — no `imagePullSecrets` equivalent.
- [ ] **CPU/memory limits** — not in manifest schema.
- [ ] **Multi-node remote stop** on `remove` — remote instances may need manual cleanup if not on leader node.

### Operations

- [ ] **Structured audit log** for `apply`/`remove`/`scale` (who changed what).
- [ ] **Metrics and alerting** (Prometheus, replica drift, proxy 5xx).
- [ ] **CLI retry flags** (`apply --wait`, `--timeout`) for leader election and reconciliation.
- [ ] **Reconciler backoff** for repeated deploy failures (avoid hot loop on bad image).
- [ ] **Official HA story** — multi-leader Raft UI/API guidance, leader stickiness for CLI.
- [ ] ** systemd unit** shipped and documented in install script for all platforms.
- [ ] **Versioned migrations** for FSM state / manifest API.

### Observability of failures

- [ ] **Surface reconciler errors in `kudo status`** (last error per app/instance).
- [ ] **Centralized log aggregation** guide (ELK, Loki) for agent JSON logs.

---

## Quick reference

| Goal | Command |
|------|---------|
| Deploy | `kudo apply -f app.yaml` |
| Check | `kudo status <name>` |
| Scale | `kudo scale <name> --replicas N` |
| Remove | `kudo remove -f app.yaml` |
| Nodes | `kudo nodes` |
| Join token | `kudo token create --ttl 24h` |

---

## Related docs

- [CLI Usage](cli-usage.md)
- [Application Manifest](application-manifest.md)
- [Agent Configuration](configuration.md)
- [Runtime Architecture](architecture.md)
- [Getting Started](getting-started.md) — shorter first-time tutorial
