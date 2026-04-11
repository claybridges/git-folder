# git-folder project context

## What this is

`git-folder` is a Go CLI tool (`cmd/git-folder/`) that manages groups of related git branches under a common prefix — e.g. `async/1`, `async/2`, `async/3`. It's a git subcommand: `git folder <command>`.

## Commands

**Porcelain (user-facing):**
- `list <folder>` — list all branches in a folder
- `increment [folder]` — create and check out the next numbered branch; errors if not on the max branch when folder is inferred
- `squash` — increment + squash all commits since trunk divergence into one
- `delete [--force] <folder>` — delete all branches in folder (with confirmation)
- `delete-upto [--force] <folder> <n>` — delete numeric branches below n
- `rename [--force] <old> <new>` — rename a folder prefix

**Plumbing:**
- `max branch <folder>` — print full name of highest-numbered branch (e.g. `async/4`)
- `max number <folder>` — print numeric suffix of highest-numbered branch

**`--force` / `-f`**: skip confirmation prompts; also allows delete/rename of the currently checked-out branch by detaching HEAD first (detach happens after confirmation, immediately before the first mutation).

## Key design decisions

- **Folder names may not contain `/`** — `my/thing` is rejected; use `my` as the folder name. Enforced via `validateFolder()` in main.go.
- **"Max branch"** is the canonical term for the highest-numbered branch in a folder.
- **`Enumerate()` uses `git for-each-ref`** (not `git branch --list`) for efficiency.
- **`branchesPreflight()`** is a pure validation function (no side effects) that checks whether branches are safe to modify — worktree conflicts fail closed, `git symbolic-ref` errors are propagated. Callers detach HEAD themselves after confirmation.
- **`squash` takes no args** — always infers from current branch, errors if not on the max branch.
- **Float suffixes supported** — `async/2.5` is valid and sorts between `async/2` and `async/3`.

## Code structure

```
cmd/git-folder/
  main.go          — all commands, flag parsing, git helpers
  main_test.go     — integration tests using real git repos in temp dirs
  USAGE.txt        — embedded usage text
internal/
  folder/
    folder.go      — IsValid, Name, Number, NumberFloat, Enumerate, LastNumber,
                     MaxBranch, CurrentBranch, CurrentFolder, DetectTrunk
  example/
    example.go     — test fixture: sets up async/ folder structure
plans/
  squash.md        — implementation notes for squash (completed)
git-folder.1.md   — man page source (pandoc → git-folder.1); **never edit git-folder.1 directly** — edit git-folder.1.md and run `go tool task man` to regenerate it
```

## Testing

Tests use real git repos created in `t.TempDir()`. `initTestRepo()` creates a repo with `main` branch + a few commits. `example.Setup()` creates the full `async/` folder structure used in many tests.

Run all CI checks: `go tool task ci`

## Commits

One-line subject only; no body, no `Co-Authored-By` trailer.

