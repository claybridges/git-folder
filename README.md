# git-folder

[![CI](https://github.com/claybridges/git-folder/actions/workflows/ci.yml/badge.svg)](https://github.com/claybridges/git-folder/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/claybridges/git-folder/branch/main/graph/badge.svg)](https://codecov.io/gh/claybridges/git-folder)
[![Go Report Card](https://goreportcard.com/badge/github.com/claybridges/git-folder)](https://goreportcard.com/report/github.com/claybridges/git-folder)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A git subcommand for managing groups of branches as a folder, like `async/1`, `async/2`, etc. Using a sequence of branches can help organize and preserve interim state during iterative development. Along that path, you can squash, rebase, and experiment freely, while keeping prior work accessible.

## Why?

When iterating on an idea, it's sometimes useful to keep prior branches around. In case of a failed experiment or botched rebase, this can spare you from lost time poring over a crowded reflog. Or sometimes, it's handy to preserve a series of commits you made during development; but you also need to squash and rebase that work to move forward. Folks handle this lots of ways: `temp`, `holdIt`, `foo4`, `TCKT-123_v11`, etc. `git folder` does that with a structured approach. 

See the Example section below for more info on this workflow.

## Install

### Homebrew (macOS)

```bash
brew install claybridges/git-folder
```

### From source

Requires Go:
```bash
go install github.com/claybridges/git-folder/cmd/git-folder@latest
```

## Usage

```
usage: git folder <command> [<options>] [<args>]

Commands:
  delete          <folder>            delete all branches in folder
  delete-upto     <folder> <n>        delete numbered branches below n
  increment       [branch]|[folder]   create and checkout next numbered branch
  list            [folder]            list branches in folder
  rename          <existing> <new>    rename folder prefix

Mostly for internal use:
  last-number     [folder]            print highest numbered branch

Options:
  --force | -f                        skip confirmation prompts, detach & override checked out branches
  --nocheckout | -n                   (increment only) create branch without checking out
```
Notes:
* When `[folder]` is omitted, defaults to folder of current branch, if applicable.
  
* `delete`, `delete-upto`, and `rename` have confirmation prompts (skipped with `--force`).
  
* `increment`
    - if default is not the highest branch, errors
    - If folder is specified (rather than using default), creates a new branch from the highest numbered branch.

* `git folder completion` provides zsh completion stubs.

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