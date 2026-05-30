# Configuration Reference

## Agent Config (`kudo.yaml`)

```yaml
node:
  name: "node-1"              # Node identifier (defaults to hostname)
  bind_addr: "0.0.0.0"        # Gossip bind address
  bind_port: 7946             # Gossip bind port
  advertise_addr: "10.0.0.1"  # Address advertised to other nodes
  data_dir: "/var/lib/kudo"   # Raft and local state directory

cluster:
  bootstrap: false            # Bootstrap a new cluster (first node only)
  join_addrs: []              # Addresses of nodes to join
  join_token: ""              # HMAC join token for authentication

api:
  grpc_port: 9090             # gRPC API port
  http_port: 8080             # HTTP management port

proxy:
  http_port: 80               # L7 reverse proxy HTTP port
  https_port: 443             # L7 reverse proxy HTTPS port

log:
  level: "info"               # Log level: debug, info, warn, error
```

### Defaults

| Field | Default |
|-------|---------|
| `node.bind_addr` | `0.0.0.0` |
| `node.bind_port` | `7946` |
| `node.data_dir` | `/var/lib/kudo` |
| `api.grpc_port` | `9090` |
| `api.http_port` | `8080` |
| `proxy.http_port` | `80` |
| `proxy.https_port` | `443` |
| `log.level` | `info` |

## Application Config

```yaml
kind: Application
name: my-app
adapter: docker          # docker | nodejs | python
replicas: 3

spec:
  image: myregistry/my-app:v1.0   # Docker: container image
  entrypoint: "npm start"         # Node.js: start command
  directory: /app                 # Working directory
  env:
    PORT: "8080"
  ports:
    - 8080
    # Or map container, host, and ingress ports explicitly:
    # - port: 8080        # container port your app listens on
    #   public: 80       # port clients use on the Kudo L7 proxy (e.g. internet-facing 80)
    #   host: 0          # optional fixed Docker host port (0 or omit = ephemeral)

routing:
  domain: api.example.com
  path: /
  ingress_port: 80       # optional; should match agent proxy.http_port for that listen port
  local_access: true     # also route localhost / 127.0.0.1 (useful for local dev)
  tls: auto                       # auto | manual | off
  algorithm: round-robin          # round-robin | least-connections
  healthcheck:
    path: /health
    interval: 10s
    timeout: 3s
    unhealthy_threshold: 3
```

### Multi-Document Files

Separate multiple applications with `---`:

```yaml
kind: Application
name: app1
adapter: docker
replicas: 2
spec:
  image: app1:latest
---
kind: Application
name: app2
adapter: docker
replicas: 1
spec:
  image: app2:latest
```

## CLI Flags

### `kudo agent`

| Flag | Description |
|------|-------------|
| `-c, --config` | Path to config file (default: `/etc/kudo/kudo.yaml`) |
| `--bootstrap` | Bootstrap a new cluster |
| `--join` | Addresses of existing nodes to join |
| `--token` | Join token for authentication |
| `--name` | Node name |

### `kudo apply`

| Flag | Description |
|------|-------------|
| `-f, --file` | Path to YAML config file (required) |

### `kudo scale`

| Flag | Description |
|------|-------------|
| `--replicas` | Target replica count (required) |

### `kudo token create`

| Flag | Description |
|------|-------------|
| `--ttl` | Token time-to-live (default: `24h`) |
