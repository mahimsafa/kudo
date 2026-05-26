# Kudo Architecture

## Overview

Kudo is a single Go binary that runs as either an **agent** (cluster participant) or a **CLI** (command issuer). All cluster nodes are peers; there is no separate control plane.

```
┌─────────────┐     gRPC      ┌─────────────────────────────────────┐
│  kudo CLI   │──────────────▶│           Kudo Agent (per node)      │
└─────────────┘               │  ┌─────────┐  ┌──────┐  ┌─────────┐  │
                              │  │  Raft   │  │Gossip│  │Executor │  │
                              │  │  (FSM)  │  │      │  │(Docker) │  │
                              │  └────┬────┘  └──┬───┘  └────┬────┘  │
                              │       │          │           │        │
                              │  ┌────▼──────────▼───────────▼────┐  │
                              │  │         Reconciler Loop        │  │
                              │  └────────────────────────────────┘  │
                              │  ┌────────────────────────────────┐  │
                              │  │      L7 Reverse Proxy          │  │
                              │  └────────────────────────────────┘  │
                              └─────────────────────────────────────┘
```

## Components

### Gossip (memberlist)

- UDP-based peer discovery and failure detection
- Each node advertises its name and address
- New nodes join via `--join` flag with seed addresses
- Events: node joined, node left, node updated

### Raft Consensus

- Replicated log for cluster state (applications, nodes, instances)
- Only the Raft leader runs the reconciler loop
- State changes applied via FSM (`Apply` on log entries)
- BoltDB for persistent log and stable store

### FSM (Finite State Machine)

Stores three entity types:

| Entity | Key Operations |
|--------|----------------|
| Application | set, delete |
| Node | set, delete |
| Instance | set, delete |

### Reconciler

Runs every 10 seconds on the Raft leader:

1. Read desired state (application replica counts)
2. Compare with actual state (running instances)
3. Emit actions: deploy (scale up) or stop (scale down)
4. Scheduler picks nodes using spread strategy with anti-affinity

### Executor & Adapters

- **Docker** (built-in): pulls images, creates containers, manages lifecycle
- **Plugins** (post-MVP): Node.js and Python adapters via gRPC over Unix sockets

### L7 Reverse Proxy

- Routes by `Host` header (domain + `/` path)
- Round-robin load balancing across backend URLs
- Returns 404 for unknown domains, 502 when no backends available

### gRPC API

| RPC | Description |
|-----|-------------|
| `Apply` | Apply YAML application config |
| `GetStatus` | Get application status and instances |
| `ListNodes` | List cluster nodes |
| `ListApplications` | List all applications |
| `ScaleApplication` | Scale app to N replicas |

## Data Flow: Deploy Application

```
CLI apply ──gRPC──▶ API Server ──Raft Apply──▶ FSM (desired state)
                                                    │
                                          Reconciler Loop (leader)
                                                    │
                                          Scheduler.PickNode()
                                                    │
                                          Executor.Deploy() ──▶ Docker
                                                    │
                                          Proxy.AddRoute() ──▶ backends
```

## Join Token Flow

```
Leader: kudo token create
  └─▶ HMAC-signed token with expiry

New node: kudo agent --join <addr> --token <token>
  └─▶ Validate HMAC + expiry before accepting join
```
