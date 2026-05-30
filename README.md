# Kudo

Kudo is a lightweight orchestration tool for deploying and managing applications across multiple VPS servers. It uses Raft consensus for cluster state, memberlist for gossip-based discovery, gRPC for inter-node communication, and a built-in L7 reverse proxy for load balancing.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/mahimsafa/kudo/main/install.sh | bash
```

Or with options:

```bash
# System-wide install of a specific version
curl -fsSL https://raw.githubusercontent.com/mahimsafa/kudo/main/install.sh | bash -s -- --system --version v0.1.0

# User-level install
curl -fsSL https://raw.githubusercontent.com/mahimsafa/kudo/main/install.sh | bash -s -- --user

# Development install (clone + build from main)
curl -fsSL https://raw.githubusercontent.com/mahimsafa/kudo/main/install.sh | bash -s -- --dev --user

# Uninstall
curl -fsSL https://raw.githubusercontent.com/mahimsafa/kudo/main/install.sh | bash -s -- --uninstall
```

## Quick Start

```bash
# Bootstrap a new cluster (first node)
kudo agent --bootstrap --name node-1

# Generate a join token (on leader)
kudo token create

# Join additional nodes
kudo agent --join <leader-ip>:7946 --token <token> --name node-2

# Deploy an application
kudo apply -f configs/examples/docker-app.yaml

# Check status
kudo status
kudo status nginx-demo

# Scale an application
kudo scale nginx-demo --replicas 3

# List cluster nodes
kudo nodes
```

## Architecture

- **Agent**: Runs on each server, participates in gossip + Raft, manages workloads
- **CLI**: Issues commands to the cluster via gRPC
- **Raft**: Consensus for application/node/instance state
- **Gossip**: Node discovery via HashiCorp memberlist
- **Executor**: Adapter-based workload lifecycle (Docker built-in)
- **Reconciler**: Control loop that scales apps to desired replica count
- **Proxy**: L7 reverse proxy with round-robin load balancing

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and documentation requirements.

## Documentation

- [Repository Layout](ARCHITECTURE.md)
- [Documentation Index](docs/index.md)
- [Getting Started](docs/getting-started.md)
- [Deploy a Node.js App (Docker)](docs/deploy-nodejs-docker.md)
- [Configuration Reference](docs/configuration.md)
- [Runtime Architecture](docs/architecture.md)

## License

MIT
