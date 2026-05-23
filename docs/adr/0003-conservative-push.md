# Push is conservative by default: confirm, honour guards, no bypass flags

When jjkit pushes a bookmark it (1) pushes a **selected** bookmark via
`jj git push --bookmark <name>`, (2) shows a confirmation dialog naming the bookmark,
target change, remote, and whether the push creates / moves / force-moves the remote,
and (3) honours jj's built-in push safety and the repo's `git.private-commits`
(`wip:`) guard — it never passes `--allow-*` / bypass flags on the user's behalf. When
jj refuses a push, jjkit surfaces jj's message rather than retrying with force.

This is a deliberate deviation worth recording because the friction is the point.
Push is outward-facing and hard to reverse, and this repo's CLAUDE.md forbids
force-push to reviewed branches and bypassing private-commit guards. A future
contributor would otherwise reasonably "improve" the UX by adding a one-key force or a
skip-confirmation default — this ADR marks that as a regression of an intended safety
constraint, not a missing feature.

## Consequences

- Force-equivalent pushes (e.g. `--change`, moving a remote bookmark to a rewritten
  change) require an explicit, separate, clearly-labelled action — never the default
  path.
- A "skip confirmation" convenience setting was rejected for MVP for the same reason.
