# Documentation Index

Use this index to find the right Kudo documentation file quickly. Prefer linking to the canonical doc instead of duplicating the same explanation in multiple places.

## Guides

| Doc | Use when you need... |
|-----|----------------------|
| [Getting Started](getting-started.md) | Installation from source, first cluster setup, basic CLI commands, and first app deployment |
| [Configuration Reference](configuration.md) | Agent config fields, application manifest fields, routing, health checks, and example YAML |
| [Deploy a Node.js App (Docker)](deploy-nodejs-docker.md) | End-to-end Node.js deployment workflow using the Docker adapter |
| [Runtime Architecture](architecture.md) | Runtime components, gossip, Raft, FSM state, reconciler flow, executor, proxy, and gRPC API |

## When To Update Docs

| Functional change | Update |
|-------------------|--------|
| Install, bootstrap, joining nodes, basic CLI usage, or first deployment flow | [Getting Started](getting-started.md) |
| CLI flags or defaults that affect setup or common operations | [Getting Started](getting-started.md) and, if config-related, [Configuration Reference](configuration.md) |
| Agent YAML fields, application manifest schema, routing fields, or health-check fields | [Configuration Reference](configuration.md) |
| Runtime architecture, component responsibilities, state flow, consensus behavior, scheduling, executor behavior, proxy routing, or gRPC semantics | [Runtime Architecture](architecture.md) |
| Node.js Docker deployment steps, assumptions, or troubleshooting | [Deploy a Node.js App (Docker)](deploy-nodejs-docker.md) |
| A new user-facing feature with no suitable existing home | Add a focused new doc in `docs/` and update this index |

## Documentation Sync Rule

When syncing docs with a feature branch, update only the minimal relevant sections. Ignore grammar-only, spelling-only, style-only, or wording-preference changes unless they are required to accurately describe changed behavior.
