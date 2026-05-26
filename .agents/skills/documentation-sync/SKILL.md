---
name: documentation-sync
description: Use when a Kudo feature branch, PR, or branch diff may require user-facing documentation updates under docs/, especially before opening or finalizing a PR.
---

# Documentation Sync

## Overview

Keep Kudo's canonical docs aligned with shipped behavior. This is not a prose polish pass: every edit should trace to an observable change in behavior, APIs, configuration, workflows, or operations.

## When to Use

- A Kudo feature branch or PR needs a docs impact check.
- A branch changes CLI, API, config/schema, defaults, routing, deployment workflows, supported features, removals, deprecations, security, status, retry, or failure behavior.
- Existing branch docs need scope, canonical-home, or link verification.

## When Not to Use

- Grammar, spelling, tone, wording preference, or style-only review.
- General code review or implementation work.
- README, CONTRIBUTING, AGENTS, ARCHITECTURE, or other non-`docs/` updates unless the user explicitly expands scope.
- Refactors, file moves, tests, fixtures, generated samples, or example image swaps that leave documented behavior accurate.

## Required Inputs

| Input | Required | Description |
|-------|----------|-------------|
| Base branch | Yes | Usually `main`; use `origin/main` if local `main` is unavailable. Ask if neither exists. |
| Branch diff | Yes | Inspect actual hunks from the confirmed base, normally `main...HEAD` or `origin/main...HEAD`. |
| Current docs | Yes | Start with [`docs/index.md`](../../../docs/index.md), then read the smallest relevant docs. |
| Summaries | No | Treat Cursor, AI, or human summaries as leads only. Verify against files. |

## Workflow

### 1. Establish the Comparison

- Check branch, status, and base availability.
- Inspect file list and relevant diff hunks directly.
- Leave unrelated dirty worktree changes untouched.
- Trust the verified diff over generated summaries.

### 2. Classify Docs Impact

Ask: "Would a user or operator make a different decision after reading the docs because of this branch?"

- Update docs for observable user/operator behavior: CLI, API semantics, config schema/defaults, ports, routing, workflows, features, architecture/data flow, security, status, retry, and failure modes.
- Do not update docs for behavior-preserving refactors, compatibility-only renames, grammar/style temptations, tests, fixtures, generated samples, or nearby stale prose unrelated to the branch.
- Reject each candidate edit that cannot be mapped to a functional diff.

### 3. Choose the Canonical Home

- Use [`docs/index.md`](../../../docs/index.md) to choose the smallest existing page.
- Add a new `docs/*.md` only when no existing page fits.
- Update `docs/index.md` when docs are added, removed, renamed, or substantially re-scoped.
- Link to canonical docs instead of duplicating explanations.

### 4. Edit Minimally

- Patch only affected sections, table rows, snippets, or links.
- Preserve accurate surrounding wording.
- Avoid whole-file rewrites, broad reorganization, and drive-by cleanup.
- Keep the diff limited to `docs/` unless scope was explicitly expanded.

### 5. Validate and Report

- Re-check diff scope and mapping from each edit to the branch diff.
- Verify Markdown links point to existing files.
- Create a PR only when explicitly requested, and only after validation.

## Decision Shortcuts

| Situation | Correct Action |
|-----------|----------------|
| Internal refactor with unchanged behavior | No docs update; explain why. |
| Functional change already accurately documented | No further edits; report coverage. |
| New doc added by branch | Ensure `docs/index.md` links it; avoid rewriting it unless inaccurate. |
| Doc removed or renamed | Update `docs/index.md` and any `docs/` links that now break. |
| README would benefit from a link | Leave README unchanged unless the user requested non-`docs/` scope. |
| User asks for style cleanup during sync | Separate it from documentation sync; keep functional edits minimal. |
| Cursor/AI summary disagrees with diff | Use the verified diff, not the summary. |

## Red Flags

- "While here, make the docs sound better."
- "The files moved, so architecture docs must change."
- "The generated summary says there is a new API."
- "README should mention this too."
- "A new docs page is enough; the index can wait."

Treat these as prompts to re-check scope, observable behavior, and `docs/index.md`.

## Bundled Resources

| Path | When to load |
|------|----------------|
| [`assets/output-template.md`](assets/output-template.md) | Whenever reporting sync results |
| [`assets/pr-body-template.md`](assets/pr-body-template.md) | Only when the user explicitly asks for a docs-only PR |
| [`references/evaluator.md`](references/evaluator.md) | When testing or changing this skill before shipping workflow updates |
| [`evals/evals.json`](evals/evals.json) | When running automated skill evals (skill-creator) |

Eval artifacts: `.agents/skills/documentation-sync-workspace/` (gitignored; sibling of this skill directory).

## Validation Checklist

- [ ] Base branch and diff were inspected directly.
- [ ] External summaries, if used, were verified against actual files.
- [ ] Each docs edit maps to a functional behavior change.
- [ ] Final diff is limited to `docs/` unless the user expanded scope.
- [ ] Changes are minimal and targeted.
- [ ] New/removed/renamed docs are reflected in `docs/index.md`.
- [ ] Markdown links point to existing files.
- [ ] PR was created only if explicitly requested.

## Common Failure Patterns

| Pitfall | Solution |
|---------|----------|
| Treating a refactor as user-facing | Require an observable behavior change before editing docs. |
| Trusting a generated summary | Verify against `git diff` and source files. |
| Editing for style while touching docs | Limit edits to functional accuracy. |
| Rewriting an entire doc | Patch only the affected section or lines. |
| Duplicating existing guidance | Link to the canonical doc selected through `docs/index.md`. |
| Adding a doc without updating the index | Update `docs/index.md` in the same docs change. |
| Fixing nearby unrelated stale prose | Leave it unless it contradicts the branch behavior. |
| Creating a PR unprompted | Stop after verification unless the user asks for PR creation. |

## Output Format

Report using [`assets/output-template.md`](assets/output-template.md). Omit empty sections only when they truly do not apply.
