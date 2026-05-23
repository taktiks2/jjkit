# jjkit

A terminal UI that makes Jujutsu (jj) operations easy — "lazygit for jj". It exists
to give a lazygit-style, always-visible multi-pane view over jj's model, and to make
the bookmark-advancement workflow frictionless.

This glossary fixes the vocabulary the tool, its code, and its docs use. Most terms
come from jj itself; the entries below pin down the ones that are easy to confuse with
their git equivalents.

## Language

### jj model

**Change**:
A unit of work identified by a stable _change id_ that survives rewrites (rebase,
amend, describe). The primary thing the Log pane lists and the user navigates.
_Avoid_: commit (see below), revision (use only when quoting jj's own output).

**Commit**:
The immutable git-level snapshot a change currently points to, identified by a commit
hash that changes whenever the change is rewritten. We say "change" for the thing the
user acts on and reserve "commit" for the underlying snapshot/hash.
_Avoid_: using "commit" and "change" interchangeably.

**Working copy** (`@`):
The change currently checked out. jj auto-commits edits to it, so there is no staging
area. "`@-`" is its parent.
_Avoid_: index, staging area, HEAD.

**Bookmark**:
A named pointer to a change — jj's equivalent of a git branch. It does **not** move
automatically when new changes are created on top of it; it must be advanced. A
bookmark may track a remote counterpart.
_Avoid_: branch, ref.

**Remote-tracking bookmark**:
The local record of where a bookmark sits on a given remote (e.g. `main@origin`).
The Bookmarks pane shows it alongside the local bookmark so divergence (local moved,
remote moved) is visible.
_Avoid_: upstream, tracking branch.

**Push**:
Sending a bookmark's target to a remote. jjkit pushes a **selected bookmark**
(`--bookmark`); pushing a bare change (`--change`, which makes jj generate a bookmark
name) is a separate, explicit action. Moving a remote bookmark to a rewritten change
is force-equivalent.
_Avoid_: publish, upload.

**Operation** / **Operation log** (oplog):
A recorded jj repo-state transition and the log of them. Undo and recovery work by
restoring an operation, not by editing changes.
_Avoid_: history (ambiguous with the change graph), reflog.

**Revset**:
A query expression that selects a set of changes (e.g. `trunk()..@`). The Log pane is
defined by a revset.

### jjkit concepts

**Advance** (a bookmark):
Move a bookmark forward to a newer change so it follows recent work. Generalises the
user's `jj tug` alias ("move the nearest ancestor bookmark to `@-`"). In jjkit the
user advances a bookmark by selecting it in the Bookmarks pane and moving it to `@`.
_Avoid_: tug (alias-specific), push (push is a separate, remote operation).

## Example dialogue

> **Dev:** The user pressed the advance key on `feature` in the Bookmarks pane.
> **Expert:** Right — that moves the `feature` _bookmark_ to point at `@`, the working
> copy. It doesn't touch the _change_ under it or its commit hash.
> **Dev:** And if they'd made three `jj new`s since, the bookmark was left behind?
> **Expert:** Yes. Bookmarks don't auto-advance, so it was still on the older change.
> Advancing fixes that. If they regret it, that's one _operation_ — undo restores the
> prior op, it doesn't rewrite the change.
