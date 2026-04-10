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
async/bigbooty
async/temp
```

I know I can safely get rid of most of those now, so I do:

```
$ git folder delete-upto async 4
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
confirm? y/N
```

After the PR merges, I can clean up the rest with:
```
$ git folder delete async
```

## Development Setup

To work on `git folder`, you should install go and go-lint. You can do that with:

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

**Note:** The dev version installs to the first of `$XDG_BIN_HOME`, `~/.local/bin`, or `~/bin`.

## Colophon

Somewhat vibe-coded with lots of Claude Code, a smidge of Gemini, with _a lot_ of opinionated direction, hand-holding, and questioning. Original `go` port based on [script version](.archive/gitBranchFolder.sh).