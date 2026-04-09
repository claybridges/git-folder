# git-folder

[![CI](https://github.com/claybridges/git-folder/actions/workflows/ci.yml/badge.svg)](https://github.com/claybridges/git-folder/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/claybridges/git-folder/branch/main/graph/badge.svg)](https://codecov.io/gh/claybridges/git-folder)
[![Go Report Card](https://goreportcard.com/badge/github.com/claybridges/git-folder)](https://goreportcard.com/report/github.com/claybridges/git-folder)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A git subcommand for managing groups of branches as a folder. Particular focus on managing sequences of numbers, like `a/1`, `a/2`, etc.

## TLDR

The tool helps preserve work during iterative development by maintaining numbered branch sequences, so you can experiment freely while keeping prior work accessible.

# Why?

When iterating on an idea, it's sometimes useful to keep prior branches around. In case of a failed experiment, this can spare you from lost time poring over a crowded reflog. Or sometimes, it's handy to preserve a series of commits you made during development; but you also need to squash and rebase that work to move forward. Folks handle this lots of ways: `temp1`, `holdIt4`, `TCKT-123_v11`, etc. `git folder` canonicalizes that, grouping git branches into folders. 

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

### Completion

`git folder completion` provides zsh completion stubs.

## Usage

```
usage: git folder <command> [<options>] [<args>]

Commands:
  delete          <folder>              delete all branches in folder
  delete-upto     <folder> <n>          delete numbered branches below n
  increment       [<folder>]            create and checkout next numbered branch
  last-number     [<folder>]            print highest numbered branch
  list            [<folder>]            list branches in folder
  rename          <existing> <new>      rename folder prefix

Options:
  --force                               skip confirmation prompts
  --nocheckout                          (increment only) create branch without checking out

Notes:
  When [<folder>] is omitted, defaults to folder of current branch, if applicable.
  
  delete, delete-upto, and rename have confirmation prompts (skipped with --force).
  
  increment errors if any other than the highest branch is used for the default.
  If folder is specified (rather than using default), creates a new branch from 
  the highest numbered branch.
```

# Example

Let's say I'm converting something to use `async`/`await`:
- first branch is `async/1`
- make 7 commits
- want to move forward, squash & rebase on `main`

With `async/1` checked out, I could
```bash
$ git folder increment
```
This creates new branch `async/2`. I'd squash & rebase that, with the confidence that the original series of commits is easily accessible.
```bash
$ git rebase -i main
```

> [!NOTE]
> For good or ill, I have my git commands _heavily_ aliased, so in use `git folder increment` is `gfi` for me. 
I could squash and rebase that on `main`, and not really worry if I screw something up. Let's say I do that over a week or two, and have gotten to a PR. I can do
```bash
$ git folder list async
async/1
async/2
async/2.5
async/3
async/4
async/bigbooty
async/temp
```
Let's say I want to get rid of most of the old stuff right now, and the rest after the merge. I could do:
```bash
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

## Development

### Setup

Install development dependencies:

```bash
brew bundle
```

### Running CI locally

Run all CI checks (tests, lint, build):

```bash
make ci
```

Other useful targets:
- `make test` - Run tests with coverage
- `make lint` - Run golangci-lint
- `make build` - Build binary
- `make clean` - Remove built artifacts
- `make help` - Show all available targets

### Toggle between brew and dev versions

To toggle between the Homebrew-installed version and a local development build:

```bash
./toggleInstallBrewOrDev.sh
```

This script:
- Detects which version is currently active
- Switches between brew and local dev versions
- Automatically builds the local version when switching to dev mode
- Displays the active version after switching

**Note:** The script installs to `$XDG_BIN_HOME`, `~/.local/bin`, or `~/bin` (whichever exists first). Ensure the chosen directory is in your PATH.





# 

Organize branches into folders like `async/1`, `async/2`, `async/3` and manage them as a group.

Short aliases: `ls`, `inc`, `rm`, `mv`.

## Example

```
$ git checkout -b feature/1
$ git folder inc
creating feature/2

$ git folder list feature
feature/1
feature/2

$ git folder rename feature experiment
rename feature/ -> experiment/:
  feature/1 -> experiment/1
  feature/2 -> experiment/2
confirm? y/N y

$ git folder delete experiment
delete all branches in folder experiment/:
  experiment/1
  experiment/2
confirm? y/N y
```

## Why


## Colophon

Vibe-coded with [Claude Code](https://claude.ai/code). Original `go` port based on [script version](.archive/gitBranchFolder.sh).
