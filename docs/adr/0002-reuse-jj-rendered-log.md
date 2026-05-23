# Render the Log pane by reusing jj's own output, not by drawing the graph ourselves

The Log pane displays jj's colored, multi-line log output as-is (`jj log --color
always` with a template), and we map cursor position back to a change to drive
selection — rather than fetching structured parent data with `--no-graph` and drawing
the commit graph ourselves.

A future reader will reasonably expect the opposite: a polished lazygit-style TUI
usually self-renders its graph for full control over selection, colour, and width.
We deliberately don't, because jjkit's display differentiator is the **always-visible
multi-pane layout**, not graph drawing. Reusing jj's renderer keeps the graph correct
and idiomatic for free and lets us reach the multi-pane MVP faster. Self-rendering
(the rejected option) buys pixel-level control we don't need yet.

## How selection works despite reusing rendered output

- We run `jj log` with a template that embeds a parseable change-id marker on each
  node so we can map screen lines → change.
- Each change keeps jj's default multi-line node (graph line + description). The
  cursor moves change-by-change and highlights all lines belonging to the selected
  change. Selecting a change drives the Diff / Bookmarks panes.

## Consequences

- We are coupled to the shape of `jj log` output; template changes or a jj output
  change can require parser updates.
- Moving to self-rendered graphs later (e.g. for custom width/colour) is a Log-pane
  rewrite. Accepted for now.
