# CLAUDE.md

## Agent skills

### Commit messages

All commits follow Conventional Commits — `<type>[(scope)][!]: <description>` (the first `jj -m` is the subject). See `docs/agents/commit-style.md`.

### Issue tracker

Issues and PRDs are tracked as GitHub issues (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Canonical triage roles map 1:1 to default label strings (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

### Go style

Go 1.21+ idioms required (slices/maps/min/max/range int/clear). See
`docs/agents/go-style.md`.
