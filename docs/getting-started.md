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

Point `demo.example.com` at any cluster node's IP on port 80. The built-in L7 proxy round-robins requests across healthy backend instances.

## Next Steps

- See [Configuration Reference](configuration.md) for all config options
- See [Architecture](architecture.md) for how components interact
