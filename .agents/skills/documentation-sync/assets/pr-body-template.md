# Documentation Sync PR Body Template

Use only when the user explicitly asked for a PR and the branch diff is documentation-only.

```markdown
## Summary
- Maps branch functional changes to documentation updates under `docs/`.

## Documentation changes
- `docs/...`: ...

## Out of scope (not in this PR)
- README / CONTRIBUTING / AGENTS / ARCHITECTURE (unless explicitly requested)
- Code or non-docs changes

## Test plan
- [ ] Diff is limited to `docs/` (and `docs/index.md` if the doc set changed)
- [ ] Each edit maps to an observable behavior change on the branch
- [ ] Markdown links resolve to existing files
```
