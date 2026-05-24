# Map screen lines to changes with an in-template sentinel marker, parsed in one pass

The Log pane runs `jj log --color always` **once** with a template that prepends a
sentinel-wrapped marker — `%%JJK%%<change-id>|<working-copy-flag>%%JJK%%` — to each
change node's first line, then keeps jj's default compact node. The parser scans each
line for the sentinel: a line that carries it begins a new change (and yields the
change-id and the `@` flag); lines without it are continuation lines of the current
change. The marker is stripped before display. This is how we realise ADR-0002's
"reuse jj's rendered output" — the screen-line → change map is embedded in the coloured
output itself.

The obvious alternative — used by lazyjj — is a **two-pass** query: one `jj log
--color always` for display plus a second `jj log --no-graph -T change_id …` for
structure, aligned line by line. We rejected it because the coloured, graph-decorated
output and a structured pass do not line up reliably — multi-line nodes, merges, and
elided (`~`) rows break a positional match — so the mapping is fragile, and it doubles
the jj invocations on every refresh. The single-pass marker keeps the mapping exact
even across multi-line nodes and keeps the parser a pure function over one byte stream,
which is exactly what the acceptance test for the screen-line → change mapping
exercises.

## Consequences

- The sentinel must not collide with real log content and must survive `--color
  always`. It does: template string literals are emitted verbatim with no colour label,
  so the marker stays plain ASCII and is found/stripped by a byte search.
- The writer (template) and reader (parser) share one `Sentinel` constant. Extending
  the payload (e.g. for the future Diff / Bookmarks panes) means changing the template
  and the parser in lockstep.
- Marker-less lines — the elided `~` row, or anything before the first node — attach to
  the preceding change. Accepted for MVP; selecting that change highlights them too.
- We stay coupled to `builtin_log_compact`'s shape (per ADR-0002); moving to a
  self-rendered graph later would replace this marker scheme.
