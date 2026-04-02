# git-folder

A git subcommand for managing numbered branch series.

Organize branches into folders like `feature/1`, `feature/2`, `feature/3` and manage them as a group.

## Install

```
go install github.com/clayb/git-folder/cmd/git-folder@latest
```

## Usage

```
git folder list <name>          # list branches in a folder
git folder increment [name]     # create next numbered branch
git folder delete <name>        # delete all branches in a folder
git folder rename <old> <new>   # rename a folder prefix
```

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

When iterating on a branch (rebasing, reworking, experimenting), it's useful to keep prior versions around as `topic/1`, `topic/2`, etc. This tool makes that workflow frictionless.
