# Go + Bubble Tea, driving jj via CLI subprocess (not Rust + jj-lib)

We build jjkit in **Go with Bubble Tea**, talking to jj by shelling out to the `jj`
CLI and parsing its output — rather than Rust with `ratatui` calling `jj-lib`
directly.

The obvious alternative was Rust, because jj is written in Rust and `jj-lib` would let
us skip output parsing entirely. We rejected it: we decided subprocess is our
**permanent** data strategy, not a stepping stone to `jj-lib`, and `jj-lib` is an
explicitly unstable API that the jj maintainers discourage external use of. Once
subprocess is permanent, Rust's main advantage disappears, and Go wins on the axes
that matter for this tool: a mature TUI ecosystem (`bubbles`/`lipgloss`) that makes
the always-visible multi-pane layout cheap to build, `tea.Cmd` + goroutines fitting
jj's slow-on-large-repos calls naturally, two existing reference implementations
(lazyjj and jjui are both Go + Bubble Tea), and faster iteration. Performance is a
non-factor — a TUI is I/O-bound on the jj subprocess, not CPU-bound.

## Consequences

- We own a parsing layer over jj's output; we lean on `jj --template` / structured
  output for stability rather than scraping human-formatted text.
- Migrating to `jj-lib` later would be a rewrite, not a refactor. We have accepted
  that.
