# AGENTS.md — Codex on jjkit

You are used here as a **read-only reviewer**. Do not modify files, and do not run
mutating commands (no `git`/`jj` write operations, no `go mod tidy`, no formatters
that write). Read, analyse, and report.

## What jjkit is

A terminal UI that makes Jujutsu (jj) operations easy — "lazygit for jj". The
implementation language is Go (a bubbletea TUI). Version control is **jj**
(a colocated git repo); do not propose `git` or `jj` mutation commands in review.

## Read these before reviewing

- `CONTEXT.md` — the domain glossary. Use its vocabulary in review comments
  (e.g. "change" vs "commit", "bookmark" vs "branch"). Don't drift to synonyms the
  glossary marks as _Avoid_.
- `docs/adr/` — architectural decisions. If a change contradicts an ADR, say so
  explicitly (e.g. "contradicts ADR-0003 conservative-push") instead of approving
  silently.
- `docs/agents/commit-style.md` — commits follow Conventional Commits.

## Review focus

- Correctness and clarity first.
- Flag violations of the glossary, the ADRs, or the commit convention.
- Surface uncertainty rather than guessing; this repo values conservative,
  reversible changes (see ADR-0003).
- Flag deviations from `docs/agents/go-style.md` (Go 1.21+ idioms).
