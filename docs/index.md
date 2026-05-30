# Documentation Index

Use this index to find the right Kudo documentation file quickly. Prefer linking to the canonical doc instead of duplicating the same explanation in multiple places.

## Guides

| Doc | Use when you need... |
|-----|----------------------|
| [Getting Started](getting-started.md) | Short tutorial: install, bootstrap, first deploy, join nodes |
| [CLI Usage](cli-usage.md) | All `kudo` commands, flags, global options, and common workflows |
| [Application Manifest](application-manifest.md) | Deployment YAML: every field, ports, routing, adapters, validation |
| [Agent Configuration](configuration.md) | Server-side `kudo.yaml` (node, cluster, API, proxy, logs) |
| [Deploy a Web Application](deploy-web-application.md) | End-to-end production deploy, logging, retries, troubleshooting, TODO gaps |
| [Deploy a Node.js App (Docker)](deploy-nodejs-docker.md) | Node.js workload example using the Docker adapter |
| [Runtime Architecture](architecture.md) | Gossip, Raft, FSM, reconciler, executor, proxy, gRPC |

## When To Update Docs

| Functional change | Update |
|-------------------|--------|
| Install, bootstrap, joining nodes, or first deployment flow | [Getting Started](getting-started.md) and/or [Deploy a Web Application](deploy-web-application.md) |
| New CLI command or flag | [CLI Usage](cli-usage.md) |
| Application manifest fields | [Application Manifest](application-manifest.md) |
| Agent config fields | [Agent Configuration](configuration.md) |
| Runtime behavior | [Runtime Architecture](architecture.md) |
| Production readiness / known gaps | [Deploy a Web Application — TODO](deploy-web-application.md#production-gaps-todo) |

## Documentation Sync Rule

When syncing docs with a feature branch, update only the minimal relevant sections. Ignore grammar-only, spelling-only, style-only, or wording-preference changes unless they are required to accurately describe changed behavior.
