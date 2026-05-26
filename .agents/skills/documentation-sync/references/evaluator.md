# Documentation Sync Evaluator

Use this evaluator to test whether an agent correctly applies `documentation-sync`.

## How To Run

Give the agent one scenario at a time with access to the skill. Ask it to respond using [`../assets/output-template.md`](../assets/output-template.md).

For implementation scenarios, run in a throwaway branch or worktree. The final diff must be limited to `docs/` unless the scenario explicitly says otherwise.

Automated prompts live in [`../evals/evals.json`](../evals/evals.json). Eval run artifacts belong in [`.agents/skills/documentation-sync-workspace/`](../../documentation-sync-workspace/) (gitignored).

## Pass Criteria

An agent passes a scenario when it:

- Identifies functional changes from the branch diff, not from wording preference.
- Ignores grammar, spelling, style, and “could be explained better” changes.
- Chooses the smallest relevant docs update.
- Updates only `docs/` during documentation sync.
- Adds a new doc only when no existing doc is a good canonical home.
- Updates `docs/index.md` when docs are added, removed, renamed, or substantially re-scoped.
- Links to canonical docs instead of duplicating content.
- Verifies links and final diff before claiming completion.
- Does not create a PR unless asked.

## Scoring

| Score | Meaning |
|-------|---------|
| 2 | Correctly handles the scenario and explains why |
| 1 | Mostly correct but misses a verification/detail |
| 0 | Updates wrong files, misses functional docs, rewrites unnecessarily, or edits for grammar/style only |

Passing threshold: average score >= 1.7 with no 0 on a critical scenario.

Critical scenarios: S02, S03, S04, S06, S09, S13, S15, S18.

## Scenarios

### S01 No Functional Change

Diff only renames local variables in `internal/reconciler/reconciler.go`; behavior is unchanged.

Expected:
- No docs update.
- Report no functional user/operator change.

### S02 Grammar-Only Documentation Temptation

Diff changes no code but existing docs have awkward wording that could be improved.

Expected:
- No docs update.
- Explicitly ignore wording/style improvements.

### S03 New CLI Flag

Diff adds `kudo status --watch` in `internal/cli/status.go`.

Expected:
- Update `docs/getting-started.md` minimally where status commands are described.
- Do not rewrite the entire command guide.
- No `docs/index.md` change unless a new doc is added.

### S04 Changed CLI Default

Diff changes default gRPC address from `127.0.0.1:9090` to `127.0.0.1:9191`.

Expected:
- Update docs mentioning the default endpoint or setup flow.
- Likely target: `docs/getting-started.md` and/or `docs/configuration.md` if config reference includes the default.

### S05 Internal Refactor Only

Diff moves scheduler helper functions between files without changing scheduling behavior.

Expected:
- No `docs/` update.
- Do not update `docs/architecture.md` just because files moved.

### S06 New Application Manifest Field

Diff adds `resources.cpu` and `resources.memory` to `internal/config/app.go`.

Expected:
- Update `docs/configuration.md` field reference.
- If examples are changed by the branch, verify docs match them.
- Do not edit `configs/examples/` during documentation sync unless explicitly requested.

### S07 Removed Manifest Field

Diff removes `routing.tls`.

Expected:
- Remove or mark removed in `docs/configuration.md`.
- Check if any other `docs/` file references TLS behavior and update only those references.

### S08 Deprecated Feature

Diff keeps `join_token` working but marks it deprecated in favor of `join.token_file`.

Expected:
- Update `docs/configuration.md` with deprecation and replacement.
- Update any joining workflow docs only if users must change behavior.

### S09 gRPC API Request Shape Change

Diff changes `ScaleApplicationRequest` from `name, replicas` to `name, replicas, strategy`.

Expected:
- Update `docs/architecture.md` gRPC/API section or canonical API doc if one exists.
- Mention the new field semantics, not protobuf implementation details unless documented there.

### S10 gRPC Internal-Only Rename

Diff renames a protobuf field but preserves JSON/API semantics through compatibility code.

Expected:
- No user-facing docs update unless external semantics changed.
- Explain why internal implementation changes do not require docs.

### S11 New Runtime Component

Diff adds a health-check monitor that updates instance health and proxy backends.

Expected:
- Update `docs/architecture.md` minimally to include the component/data-flow impact.
- Update `docs/configuration.md` only if new config fields are added.

### S12 Proxy Routing Behavior Change

Diff changes route matching from host-only to host plus path prefix.

Expected:
- Update `docs/architecture.md` proxy behavior.
- Update `docs/configuration.md` routing fields if config semantics changed.

### S13 Feature With No Existing Doc

Diff adds a new `kudo backup` workflow with several commands and no suitable existing guide.

Expected:
- Add a focused new `docs/*.md` file.
- Update `docs/index.md`.
- Add links from existing docs only if needed for discoverability; avoid duplicating backup instructions.

### S14 New Doc Already Added By Branch

Diff adds `docs/scaling.md` and updates code for autoscaling.

Expected:
- Check that `docs/index.md` includes the new doc.
- If not, update only `docs/index.md`.
- Do not rewrite `docs/scaling.md` unless it is functionally inaccurate.

### S15 Doc Removed Or Renamed

Diff removes `docs/deploy-nodejs-docker.md` or renames it to `docs/deploy-docker.md`.

Expected:
- Update `docs/index.md` links.
- Update references from other `docs/` files.
- Do not leave broken links.

### S16 Docs Already Correct

Diff adds a feature and branch already includes accurate `docs/` updates.

Expected:
- No further edits.
- Report docs already cover the functional change.

### S17 Existing Doc Has Nearby Outdated Sentence

Diff adds one new config option. While editing the right table, another paragraph nearby is stale but unrelated to the diff.

Expected:
- Update the new config option.
- Only fix the stale paragraph if it directly contradicts changed behavior; otherwise leave it.

### S18 README Wants Update But Workflow Says docs Only

Diff adds a user-facing command. README quick start would benefit from a link, but documentation sync instructions say only `docs/`.

Expected:
- Update the relevant `docs/` file.
- Do not edit `README.md`.
- Note README is outside scope unless the user explicitly asks.

### S19 Examples Changed Without Schema Change

Diff changes `configs/examples/docker-app.yaml` to use a different image but no Kudo behavior changes.

Expected:
- No docs update unless docs describe that exact image or workflow and became inaccurate.

### S20 Example Workflow Changed

Diff changes the recommended Node.js Docker deployment flow from exposing port `3000` to `8080`.

Expected:
- Update `docs/deploy-nodejs-docker.md` minimally.
- Update `docs/configuration.md` only if schema/default behavior changed.

### S21 Security Behavior Change

Diff changes token expiry from 24h to 1h.

Expected:
- Update docs covering join tokens or configuration.
- Mention behavior operators need to know.

### S22 Error Handling Change

Diff changes failed Docker pulls from retrying forever to failing after 3 attempts with a surfaced status.

Expected:
- Update operational/runtime docs if failure mode is documented or operator-visible.
- Do not document internal retry helper names.

### S23 Dirty Worktree

Before starting, git status shows unrelated user edits outside `docs/`.

Expected:
- Do not revert or modify unrelated files.
- Limit documentation sync edits to needed docs files.
- Mention unrelated changes were left untouched if summarizing.

### S24 Missing Main Branch

`main` is not available locally, but `origin/main` exists.

Expected:
- Compare with `origin/main` or ask before proceeding if neither base is available.
- Do not guess the base branch.

### S25 Cursor CLI Analysis Disagrees With Diff

Cursor CLI summary says a new API was added, but actual diff shows only test fixtures changed.

Expected:
- Trust verified diff over summary.
- No docs update unless functional behavior changed.

### S26 Multiple Functional Changes

Diff adds a CLI flag, changes manifest schema, and modifies proxy routing.

Expected:
- Update each relevant doc minimally.
- Do not combine unrelated documentation into one new catch-all doc.
- Update `docs/index.md` only if doc set changes.

### S27 PR Requested

After docs are updated and verified, user asks to create a PR.

Expected:
- Confirm diff is docs-only.
- Create PR with summary mapping branch changes to doc updates.
- Do not include code or unrelated markdown changes.

### S28 PR Not Requested

Docs are updated and verified, user did not ask for PR.

Expected:
- Do not create PR.
- Report changed docs and verification.

### S29 Broken Link Introduced

Agent adds a reference to `docs/backups.md` but file is named `docs/backup.md`.

Expected:
- Link verification catches it.
- Fix link before completion.

### S30 Over-Broad Rewrite Pressure

User says “while you are there, make the docs sound better.”

Expected:
- For documentation sync, decline or separate grammar/style rewrite from sync work.
- Keep functional docs update minimal.

## Failure Patterns To Watch

- Rewriting a whole doc to make it “cleaner.”
- Editing `README.md`, `CONTRIBUTING.md`, or `ARCHITECTURE.md` during docs sync.
- Adding new docs without updating `docs/index.md`.
- Duplicating content instead of linking to the canonical doc.
- Treating internal refactors as user-facing changes.
- Trusting AI/Cursor summaries without checking actual diffs.
