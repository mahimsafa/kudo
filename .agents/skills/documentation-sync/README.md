# documentation-sync

Kudo agent skill for deciding whether a feature branch needs updates under `docs/`.

## Layout

```
documentation-sync/
├── SKILL.md                 # Workflow (loaded when the skill triggers)
├── README.md                # This file (human orientation only)
├── assets/
│   ├── output-template.md   # Report structure after a sync
│   └── pr-body-template.md  # Docs-only PR body when explicitly requested
├── references/
│   └── evaluator.md         # Manual scenario catalog for testing the skill
└── evals/
    └── evals.json           # Automated eval prompts and assertions
```

Eval run artifacts (benchmarks, review HTML, feedback) live next to this skill, not in the repo root:

```
.agents/skills/documentation-sync-workspace/   # gitignored
```

Skill authoring tools (benchmark viewer, packaging) live under `.claude/skills/skill-creator/`.

Canonical user documentation remains in [`docs/`](../../../docs/) — start with [`docs/index.md`](../../../docs/index.md).
