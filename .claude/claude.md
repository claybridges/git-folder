# git-folder project context

## What this is

`git-folder` is a Go CLI tool (`cmd/git-folder/`) that manages groups of related git branches under a common prefix — e.g. `async/1`, `async/2`, `async/3`. It's a git subcommand: `git folder <command>`.

## Commands

See `git-folder.1.md` — full porcelain/plumbing docs, including `--force` semantics.

## Key design decisions

- **Folder names may not contain `/`** — `my/thing` is rejected; use `my` as the folder name. Enforced via `validateFolder()` in main.go.
- **"Max branch"** is the canonical term for the highest-numbered branch in a folder.
- **`Enumerate()` uses `git for-each-ref`** (not `git branch --list`) for efficiency.
- **`branchesPreflight()`** is a pure validation function (no side effects) that checks whether branches are safe to modify — worktree conflicts fail closed, `git symbolic-ref` errors are propagated. Callers detach HEAD themselves after confirmation.
- **`squash` takes no args** — always infers from current branch, errors if not on the max branch.
- **Float suffixes supported** — `async/2.5` is valid and sorts between `async/2` and `async/3`.

## Man page

**Never edit `git-folder.1` directly** — edit `git-folder.1.md` and run `go tool task man` to regenerate it.

## Testing

Run all CI checks: `go tool task ci`

## Commits

One-line subject only; no body, no `Co-Authored-By` trailer.
