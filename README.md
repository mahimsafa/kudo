# Kudo

Kudo is a lightweight orchestration tool for deploying and managing applications across multiple VPS servers. It uses Raft consensus for cluster state, memberlist for gossip-based discovery, gRPC for inter-node communication, and a built-in L7 reverse proxy for load balancing.

## Quick Start

```bash
# Build
make build

# Bootstrap a new cluster (first node)
./bin/kudo agent --bootstrap --name node-1

# Generate a join token (on leader)
./bin/kudo token create

# Join additional nodes
./bin/kudo agent --join <leader-ip>:7946 --token <token> --name node-2

# Deploy an application
./bin/kudo apply -f configs/examples/docker-app.yaml

# Check status
./bin/kudo status
./bin/kudo status nginx-demo

# Scale an application
./bin/kudo scale nginx-demo --replicas 3

# List cluster nodes
./bin/kudo nodes
```

## Architecture

- **Agent**: Runs on each server, participates in gossip + Raft, manages workloads
- **CLI**: Issues commands to the cluster via gRPC
- **Raft**: Consensus for application/node/instance state
- **Gossip**: Node discovery via HashiCorp memberlist
- **Executor**: Adapter-based workload lifecycle (Docker built-in)
- **Reconciler**: Control loop that scales apps to desired replica count
- **Proxy**: L7 reverse proxy with round-robin load balancing

## Documentation

- [Getting Started](docs/getting-started.md)
- [Deploy a Node.js App (Docker)](docs/deploy-nodejs-docker.md)
- [Configuration Reference](docs/configuration.md)
- [Architecture](docs/architecture.md)

## License

MIT
