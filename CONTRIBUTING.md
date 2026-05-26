# Contributing to Kudo

Thank you for your interest in contributing to Kudo. This guide covers how to set up a development environment, make changes, and what to include in a pull request—including **required documentation updates** when your change affects users or operators.

## Before you start

- Read [ARCHITECTURE.md](ARCHITECTURE.md) to understand where code lives in the repository.
- Read [docs/architecture.md](docs/architecture.md) for how runtime components interact.
- Open an issue for large or ambiguous changes so maintainers can align on approach before you invest significant effort.

## Development setup

### Prerequisites

- **Go 1.25+** (see `go.mod` for the exact version)
- **Docker** (for testing the Docker executor and example workloads)
- **protoc** + Go plugins (only if you change `internal/api/proto/kudo.proto`)
- **golangci-lint** (optional locally; used by `make lint`)

### Build and run

```bash
git clone https://github.com/mahimsafa/kudo.git
cd kudo
make build
./bin/kudo --help
```

Run the agent locally for manual testing:

```bash
./bin/kudo agent --bootstrap --name dev-node-1
```

In another terminal, use the CLI against the local gRPC endpoint (default `127.0.0.1:9090`):

```bash
./bin/kudo apply -f configs/examples/docker-app.yaml
./bin/kudo status
```

See [docs/getting-started.md](docs/getting-started.md) for a full cluster walkthrough.

## Making changes

### Where to put code

| Change type | Location |
|-------------|----------|
| New CLI command or flags | `internal/cli/` |
| Agent lifecycle / wiring | `internal/agent/` |
| Cluster state (Raft FSM) | `internal/cluster/state/` |
| Discovery / Raft transport | `internal/cluster/gossip/`, `internal/cluster/raft/` |
| gRPC API | `internal/api/`, `internal/api/proto/kudo.proto` |
| Workload runtime | `internal/executor/`, `internal/executor/docker/` |
| Scheduling / reconciliation | `internal/scheduler/`, `internal/reconciler/` |
| Reverse proxy | `internal/proxy/` |
| Config parsing | `internal/config/` |
| Join tokens | `internal/auth/` |

Follow existing patterns in the package you are editing. Keep changes focused; avoid unrelated refactors in the same pull request.

### Coding conventions

- Match the style of surrounding Go code (formatting, naming, error wrapping).
- Prefer small, testable functions over large monoliths.
- Return wrapped errors: `fmt.Errorf("context: %w", err)`.
- Add or update unit tests in `*_test.go` files next to the code you change.
- Do not commit generated secrets, `.env` files, or local data directories.

### gRPC / protobuf changes

If you modify `internal/api/proto/kudo.proto`:

1. Regenerate stubs: `make proto`
2. Commit both the `.proto` and generated `*.pb.go` files.
3. Update CLI and server implementations that use the new or changed RPCs.
4. **Update documentation** (see below)—API changes are always user-facing.

### Example manifests

When you add or change application YAML fields, update:

- [`internal/config/app.go`](internal/config/app.go) (parsing and validation)
- [`configs/examples/`](configs/examples/) (at least one example that demonstrates the new behavior)
- [docs/configuration.md](docs/configuration.md) (field reference)

## Testing

Before opening a pull request, run:

```bash
make test    # unit tests with race detector
make build   # confirm the binary compiles
```

If you have `golangci-lint` installed:

```bash
make lint
```

Add tests for new behavior and bug fixes. Tests should cover real logic, not trivial getters or one-line wrappers.

## Documentation requirements

**If your contribution changes behavior that users or operators need to know about, you must update documentation under [`docs/`](docs/).** Documentation updates belong in the same pull request as the code change—not in a follow-up.

Pull requests that change user-facing behavior without matching `docs/` updates will be asked to add them before merge.

### When documentation is required

Update `docs/` when you change any of the following:

- CLI commands, flags, or default values
- Agent or application YAML configuration fields
- Cluster join / token / security behavior
- Deployment, scaling, or routing behavior
- gRPC API (RPCs, request/response shapes, semantics)
- Prerequisites, ports, or operational runbooks
- Error messages or failure modes that operators must handle

### Which file to update

Use [docs/index.md](docs/index.md) to find the canonical documentation home before adding or editing docs.

| Your change affects… | Update this doc |
|----------------------|-----------------|
| Install, first cluster, basic commands | [docs/getting-started.md](docs/getting-started.md) |
| Agent or application YAML fields | [docs/configuration.md](docs/configuration.md) |
| Runtime components, data flow, consensus | [docs/architecture.md](docs/architecture.md) |
| End-to-end app deployment (e.g. Node.js) | [docs/deploy-nodejs-docker.md](docs/deploy-nodejs-docker.md) |

If a change spans multiple areas, update every relevant doc. Keep wording consistent with existing guides.

### Repository layout vs user docs

| Document | Purpose | Update when… |
|----------|---------|--------------|
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | Codebase layout, packages, build flow | You add/remove/rename packages, move entry points, or change build/proto workflow |
| [`docs/*.md`](docs/) | How to install, configure, and operate Kudo | Behavior, APIs, or config visible to users changes |
| [`README.md`](README.md) | Project overview and quick links | New top-level commands, install path changes, or new primary doc links |
| [`configs/examples/`](configs/examples/) | Runnable YAML samples | Application manifest schema or common usage patterns change |

**Rule of thumb:** `docs/` answers *how do I use Kudo?*; `ARCHITECTURE.md` answers *where is the code?*

### Documentation checklist

Before submitting your PR, confirm:

- [ ] User-facing behavior changes are reflected in at least one file under `docs/`
- [ ] New or changed YAML fields appear in `docs/configuration.md` and `configs/examples/` if applicable
- [ ] New CLI commands or flags are documented in the relevant guide (usually `getting-started.md` or `configuration.md`)
- [ ] Runtime or component changes are described in `docs/architecture.md` when they affect system design
- [ ] `ARCHITECTURE.md` is updated if you added packages or changed repository structure
- [ ] `README.md` quick start or doc links still match reality (if you changed primary workflows)

## Pull request process

1. Fork the repository and create a branch from `main`.
2. Make your changes with tests and **documentation** as described above.
3. Write a clear PR description:
   - **What** changed and **why**
   - **How** to test (commands run, expected outcome)
   - **Docs:** list which `docs/` files you updated (or state “no user-facing changes”)
4. Ensure CI checks pass (when configured on the repository).
5. Address review feedback; keep the branch up to date with `main` if requested.

### PR title and commits

Use concise, descriptive titles. Examples:

- `fix: reconciler skips stopped instances on scale-down`
- `feat: add health check interval to application routing`
- `docs: document join token expiry in configuration guide`

Squashing is fine; the final commit message on merge should still describe the change clearly.

## Reporting bugs

Include:

- Kudo version or commit SHA
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (agent, Docker) with secrets redacted

## License

By contributing, you agree that your contributions will be licensed under the MIT License used by this project (see [README.md](README.md)).
