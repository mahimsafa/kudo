# Getting Started with Kudo

This guide walks you through installing Kudo, bootstrapping a cluster, deploying your first application, and adding a second node.

## Prerequisites

- Go 1.22+ (for building from source)
- Docker (for container workloads)
- Linux or macOS VPS servers with network connectivity

## Install

```bash
git clone https://github.com/mahimsafa/kudo.git
cd kudo
make build
sudo cp bin/kudo /usr/local/bin/
```

## Bootstrap the First Node

Create an agent config at `/etc/kudo/kudo.yaml` (optional — defaults work for local testing):

```yaml
node:
  name: "node-1"
  bind_addr: "0.0.0.0"
  bind_port: 7946
  data_dir: "/var/lib/kudo"

cluster:
  bootstrap: true

api:
  grpc_port: 9090

proxy:
  http_port: 80

log:
  level: "info"
```

Start the agent:

```bash
kudo agent --bootstrap --name node-1
```

Or without a config file:

```bash
kudo agent --bootstrap --name node-1
```

## Generate a Join Token

On the bootstrap node:

```bash
kudo token create --ttl 24h
```

Save the token output. New nodes will need it to authenticate when joining.

## Deploy Your First Application

Create `my-app.yaml`:

```yaml
kind: Application
name: nginx-demo
adapter: docker
replicas: 2
spec:
  image: nginx:alpine
  ports:
    - 80
routing:
  domain: demo.example.com
  algorithm: round-robin
```

Apply it:

```bash
kudo apply -f my-app.yaml
```

Check status:

```bash
kudo status nginx-demo
```

## Remove a deployment

Use the same YAML file (or a manifest listing the apps to tear down). Kudo stops local containers, deletes cluster state, and clears proxy routes when configured.

```bash
kudo remove -f my-app.yaml
```

Removal is **all-or-nothing**: if another application still uses the same `routing.domain` + `routing.path`, nothing is removed and the CLI prints which app blocks teardown. Include every app that shares that route in the same file (for example both documents in a multi-app manifest) to remove them together.

If a container was deleted manually, `kudo remove` prints warnings and asks for confirmation before removing cluster state and routes. Pass `-y` to continue without prompting.

## Add a Second Node

On the new server:

```bash
kudo agent --join <node-1-ip>:7946 --token <join-token> --name node-2
```

Verify the cluster:

```bash
kudo nodes
```

## Verify Load Balancing

The Kudo agent runs an L7 reverse proxy (port **8088** in local dev without a config file, or **80** when configured). It forwards to each replica's Docker-published address; replicas do **not** all bind host port 80 directly.

After `apply`, check agent logs for `application reachable via L7 proxy` URLs, or use:

```bash
# With local_access: true in routing (see configs/examples/docker-app.yaml)
curl -v http://127.0.0.1:8088/

# Or send the routing domain as Host
curl -v -H 'Host: demo.example.com' http://127.0.0.1:8088/
```

To expose an app on container port **8080** at public port **80**, see `configs/examples/docker-app-public-port.yaml` and set the agent `proxy.http_port` to `80` (requires permission to bind port 80).

Point `demo.example.com` at any cluster node's IP on the proxy port. The proxy round-robins across healthy backend instances.

## Next Steps

- [CLI Usage](cli-usage.md) — all commands and flags
- [Application Manifest](application-manifest.md) — deployment YAML reference
- [Deploy a Web Application](deploy-web-application.md) — production end-to-end guide
- [Agent Configuration](configuration.md) — `kudo.yaml` per server
- [Runtime Architecture](architecture.md) — how components interact
