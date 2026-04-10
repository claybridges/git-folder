# `git folder squash` — implementation plan (completed)

Implemented in the `squash` branch. Kept here for reference.

## What it does

`git folder squash` = increment + squash all commits since divergence from trunk.

1. Bail if working tree or index is dirty
2. Detect trunk: `DetectTrunk()`
3. Find merge base: `git merge-base HEAD <trunk>`
4. Find first commit after split: `git rev-list --reverse --topo-order --ancestry-path <merge-base>..HEAD`
5. Count commits after first; if 0, error "nothing to squash" (before creating branch)
6. Call `cmdIncrement(nil)` — validates we're on the max branch, creates next branch
7. Collect messages: `git log --format=%B <first-commit>..HEAD` (strip blank lines)
8. Get first commit's message separately
9. `git reset --soft <first-commit>`
10. `git commit --amend -m "<first-commit-msg>\n\n<collected-messages>"`

## Files modified

- `internal/folder/folder.go` — added `DetectTrunk()`, `MaxBranch()`, `CurrentBranch()`; switched `Enumerate()` to `git for-each-ref`
- `cmd/git-folder/main.go` — added `cmdSquash()`, `cmdMax()`, `gitOutput()` helper; wired up in switch
- `cmd/git-folder/USAGE.txt` — added `git folder squash`
- `cmd/git-folder/main_test.go` — added `TestCmdSquash`, `TestCmdSquashNothingToSquash`, `TestDetectTrunk`, `TestCmdMax*`
- `git-folder.1.md` — added squash to SYNOPSIS and COMMANDS; porcelain/plumbing sections
