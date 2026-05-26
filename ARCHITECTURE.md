# Kudo Repository Layout

Kudo is a lightweight multi-node orchestrator written in Go. It uses Raft consensus for replicated cluster state, HashiCorp memberlist for gossip-based node discovery, a pluggable executor (Docker built-in) for workload lifecycle, and an L7 reverse proxy for load balancing.

For install instructions and a quick start, see [README.md](README.md). For how the runtime components interact (gossip, Raft FSM, reconciler loop, deploy data flow), see [Runtime Architecture](docs/architecture.md).

## Repository layout

```
kudo/
├── cmd/kudo/              # Binary entry point
├── internal/              # All application logic
│   ├── agent/
│   ├── api/
│   ├── auth/
│   ├── cli/
│   ├── cluster/
│   │   ├── gossip/
│   │   ├── raft/
│   │   └── state/
│   ├── config/
│   ├── executor/
│   │   └── docker/
│   ├── log/
│   ├── proxy/
│   ├── reconciler/
│   └── scheduler/
├── configs/examples/      # Sample application YAML manifests
├── docs/                  # User-facing guides
├── plugins/               # Placeholder dirs for future gRPC adapters
├── Makefile
├── go.mod
└── go.sum
```

Build output goes to `bin/` (gitignored). Tooling directories (`.cursor`, `.opencode`) are local IDE metadata and not part of the application.

| Path | Purpose |
|------|---------|
| [`cmd/kudo/`](cmd/kudo/) | `main()` — delegates to `internal/cli` |
| [`internal/`](internal/) | All application logic (not importable externally) |
| [`configs/examples/`](configs/examples/) | Sample application YAML manifests |
| [`docs/`](docs/) | User-facing guides |
| [`plugins/`](plugins/) | Placeholder dirs for future gRPC adapters (`nodejs-adapter`, `python-adapter`; currently empty) |
| [`Makefile`](Makefile) | `build`, `test`, `clean`, `lint`, `proto` targets |
| [`go.mod`](go.mod) / [`go.sum`](go.sum) | Module dependencies (HashiCorp raft/memberlist, Docker SDK, Cobra, gRPC) |

## `internal/` package map

All production code lives under `internal/`. Unit tests are colocated as `*_test.go` files in each package.

| Package | Role | Key files |
|---------|------|-----------|
| [`internal/cli`](internal/cli/) | Cobra commands: `agent`, `apply`, `status`, `scale`, `nodes`, `token`; shared gRPC client | `root.go`, `agent.go`, `apply.go`, `status.go`, `scale.go`, `nodes.go`, `token.go`, `grpc.go` |
| [`internal/agent`](internal/agent/) | Wires gossip, Raft, API server, executor, proxy, and reconciler on startup | `agent.go` |
| [`internal/api`](internal/api/) | gRPC server implementing cluster RPCs | `server.go`, [`proto/kudo.proto`](internal/api/proto/kudo.proto), generated `*.pb.go` |
| [`internal/cluster/gossip`](internal/cluster/gossip/) | HashiCorp memberlist discovery and failure detection | `gossip.go`, `delegate.go` |
| [`internal/cluster/raft`](internal/cluster/raft/) | Raft node with BoltDB-backed log and stable store | `raft.go` |
| [`internal/cluster/state`](internal/cluster/state/) | FSM holding applications, nodes, and instances | `fsm.go` |
| [`internal/config`](internal/config/) | Agent YAML config and application manifest parsing | `config.go`, `app.go` |
| [`internal/auth`](internal/auth/) | HMAC-signed join tokens | `token.go` |
| [`internal/executor`](internal/executor/) | Adapter registry; dispatches `Deploy` and `Stop` | `executor.go`, `adapter.go` |
| [`internal/executor/docker`](internal/executor/docker/) | Built-in Docker adapter (pull, create, lifecycle) | `docker.go` |
| [`internal/reconciler`](internal/reconciler/) | Leader-only control loop comparing desired vs actual replicas | `reconciler.go` |
| [`internal/scheduler`](internal/scheduler/) | Node placement (spread strategy with anti-affinity) | `scheduler.go` |
| [`internal/proxy`](internal/proxy/) | L7 reverse proxy (Host-based routing, round-robin) | `proxy.go` |
| [`internal/log`](internal/log/) | Zap logger setup | `logger.go` |

## Entry point and build

[`cmd/kudo/main.go`](cmd/kudo/main.go) calls `cli.Execute()`, which runs the Cobra command tree registered in `internal/cli`.

| Command | Description |
|---------|-------------|
| `make build` | Compiles `bin/kudo` from `./cmd/kudo` |
| `make test` | Runs `go test ./...` with race detector |
| `make proto` | Regenerates gRPC stubs from `internal/api/proto/kudo.proto` |
| `make clean` | Removes `bin/` |

## Configuration artifacts

**Agent config** — YAML file loaded by the agent at startup. Struct definitions live in [`internal/config/config.go`](internal/config/config.go). See [Configuration Reference](docs/configuration.md) for field descriptions.

**Application manifests** — Declarative app definitions with `kind: Application`. Examples in [`configs/examples/`](configs/examples/):

| File | Description |
|------|-------------|
| `docker-app.yaml` | Single Docker app with routing and health checks |
| `nodejs-app.yaml` | Node.js workload example |
| `multi-app.yaml` | Multiple applications in one file |

Parsed by [`internal/config/app.go`](internal/config/app.go) and applied to the cluster via `kudo apply`.

## Documentation

| Doc | Contents |
|-----|----------|
| [getting-started.md](docs/getting-started.md) | First cluster setup and basic commands |
| [configuration.md](docs/configuration.md) | Agent and application YAML reference |
| [deploy-nodejs-docker.md](docs/deploy-nodejs-docker.md) | End-to-end Node.js deployment walkthrough |
| [architecture.md](docs/architecture.md) | Runtime component design and data flows |

For contribution guidelines and when to update these docs, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Code organization

Dependency direction keeps concerns separated:

```
CLI  ──gRPC──▶  API Server  ──Raft Apply──▶  FSM (state)
                                              │
Agent ──▶ gossip + raft + reconciler ──▶ scheduler + executor + proxy
```

- The **CLI** talks to the cluster only through the gRPC client in `internal/cli/grpc.go`. It does not import agent or cluster packages directly.
- The **agent** owns the long-running process: it starts gossip and Raft, registers the Docker executor, runs the reconciler on the Raft leader, and serves the gRPC API and reverse proxy.
- **Cluster state** mutations flow through Raft log entries applied to the FSM; the reconciler reads that state and drives the executor and proxy.

## Plugins (future)

[`plugins/`](plugins/) contains empty placeholder directories for post-MVP gRPC adapters (`nodejs-adapter`, `python-adapter`). These will run as separate processes communicating with the agent over Unix sockets, following the same `internal/executor.Adapter` interface used by the Docker adapter.
