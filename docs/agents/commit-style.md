# Commit messages: Conventional Commits

All commit descriptions in this repo follow [Conventional Commits](https://www.conventionalcommits.org/).

## Format

```
<type>[(scope)][!]: <description>

[optional body]
```

The **first `-m`** you pass to `jj` is the subject and must match this form. Extra
`-m` flags become the body and are free text.

- **type** — one of: `feat` `fix` `docs` `style` `refactor` `perf` `test` `build` `ci` `chore` `revert` `wip`
- **scope** — optional, in parens, e.g. `fix(parser):`
- **!** — optional, marks a breaking change, e.g. `feat!:` / `feat(api)!:`
- **description** — imperative, lower-case, no trailing period; keep the subject ≤ ~72 chars

`wip:` is project-specific: use it for in-progress commits (they are push-guarded via
`git.private-commits`). Re-`describe` to a real type before the commit is ready to push.

## Examples (jj)

```sh
jj describe -m "feat(parser): handle nil node"
jj commit   -m "fix: stop panic on empty input"
jj new main -m "chore: scaffold feature branch"
jj describe -m "feat!: drop v1 config" -m "The v1 loader is removed; migrate to v2."
```

## Enforcement

This is a convention agents follow by reading this doc and `CLAUDE.md` — there is no
local hook (jj has no `commit-msg` hook, so client-side checks aren't built in).

If hard enforcement is ever needed, add a GitHub Actions check as the backstop for
humans and other tools: lint the PR title (squash merge) or every commit
(merge / rebase).
