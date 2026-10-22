# git-folder

[![CI](https://github.com/claybridges/git-folder/actions/workflows/ci.yml/badge.svg)](https://github.com/claybridges/git-folder/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/claybridges/git-folder/branch/main/graph/badge.svg)](https://codecov.io/gh/claybridges/git-folder)
[![Go Report Card](https://goreportcard.com/badge/github.com/claybridges/git-folder)](https://goreportcard.com/report/github.com/claybridges/git-folder)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A git subcommand for creating and managing folders of related branches, with an emphasis on iterative sequences — `a/1`, `a/2`, `a/3`, etc.

## Why?

Squashing, merging, and rebasing destroy context. That's usually fine — until it isn't. Developers already work around this with ad hoc branches — `save3`, `omg-works-here`, `dontDeleteYet`. Giving this structure as a numbered branch sequence preserves your iterative context — a branch time machine for experimentation.

## Install

### Homebrew (macOS)

```bash
brew install claybridges/tap/git-folder
```

### From source

Requires Go:
```bash
go install github.com/claybridges/git-folder/cmd/git-folder@latest
```

### Shell completion (zsh)

```zsh
git folder completion > ~/.zsh/completions/_git-folder
```

## Usage

See the [man page](git-folder.1.md) for full command reference.

## Example

Let's say I'm converting something to use `async`/`await`:
- first branch is `async/1`
- I make 7 commits
- Now it's time to squash & rebase on `main` so I can move forward

With `async/1` checked out, I could
```bash
$ git folder increment
```
This creates new branch `async/2`. I'd squash & rebase that, with the confidence that the original series of commits is easily accessible.
```bash
$ git rebase -i main
```

After iterating on that for a week, I've gotten to a PR. Listing my `async` branches, I get:

```
$ git folder list async
async/1
async/2
async/2.5
async/3
async/4
async/5
async/bigbooty
async/temp
```

I know I can safely get rid of most of those now, so I do:

```
$ git folder delete --before 4 async
keep:
  async/4
  async/5
  async/bigbooty
  async/temp

delete:
  async/1
  async/2
  async/2.5
  async/3
confirm? [yN]
```

After the PR merges, I can clean up the rest with:
```
$ git folder delete async
```

## Agents and worktrees

Agents rewrite history: rebase, squash, replay. That's fine until one eats a commit you needed. Numbered branches make it harmless. `git folder increment` first, point the rewrite at the new N, and the original stays put.

Two things that actually happened:

- `git replay --onto main base..widget/2` rewrites `widget/2` in place. The original was gone until I reset it from the log. Now: `git folder increment`, then replay `widget/3`.
- A sweep rebased seven branches onto main (`parser/5` → `parser/6`, and so on). `mic/3` hit a conflict, so `mic/4` stayed at the old tip and `mic/3` was never touched.

One worktree and one folder per agent (`login/1`, `search/1`), landing fast-forward only. Along the way:

- `delete` and `rename` refuse branches checked out in any worktree
- `increment <folder>` branches from the folder's max branch, so an agent can start from the right place anywhere
- when the PRs land, `git folder delete login search` cleans up both folders at once

## Development Setup

To work on `git folder`, you should install `go`, `golangci-lint`, and `pandoc` (needed by `go tool task man`). You can do that with:

```
brew bundle
```

To run all CI checks (tests, lint, build):

```
go tool task ci
```

You can switch your local `git folder` commands between the brew install & a local development version with these:

```
go tool task use-dev
go tool task use-brew
```

> [!NOTE]
> The dev version installs to the first of `$XDG_BIN_HOME`, `~/.local/bin`, or `~/bin`. 

### Git hook for `main`

To rebuild the dev version automatically whenever local `main` moves (`git pull`, `git branch -f main`, etc.), install the hook:

```
go tool task main-hook:install
```

This symlinks `./scripts/hooks/reference-transaction` into this repo's `hooks` dir.

To remove it:

```
go tool task main-hook:uninstall
```

`use-brew` automatically uninstalls the hook.

## Colophon

Somewhat vibe-coded with lots of Claude Code, a smidge of Gemini, with _a lot_ of opinionated direction, hand-holding, and questioning. Original `go` port based on [script version](.archive/gitBranchFolder.sh).