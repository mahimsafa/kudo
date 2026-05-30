# Agent Configuration Reference

Server-side settings for `kudo agent`. Application deployment YAML is documented separately in [Application Manifest](application-manifest.md).

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

## Application manifests

Deployment YAML (`kind: Application`, `spec`, `routing`, ports, adapters) is covered in [Application Manifest](application-manifest.md).

CLI commands and flags are in [CLI Usage](cli-usage.md).

## Related docs

- [Application Manifest](application-manifest.md)
- [CLI Usage](cli-usage.md)
- [Deploy a Web Application](deploy-web-application.md)
