% GIT-FOLDER(1) Git Manual
% Clay Bridges
% 2026-04-09

# NAME

git-folder - manage groups of git branches as folders

# SYNOPSIS

```
git folder list <folder>
git folder increment [<folder>]
git folder squash
git folder delete [--force] <folder>
git folder delete-upto [--force] <folder> <n>
git folder rename [--force] <old> <new>
git folder max branch <folder>
git folder max number <folder>
```

# DESCRIPTION

**git-folder** manages groups of related branches under a common prefix, like
`a/1`, `a/2`, `a/3`, etc.

# HIGH-LEVEL COMMANDS (PORCELAIN)

**list** *\<folder\>*
:   List all branches in the folder.

**increment** *[folder]*
:   Create and checkout the next numbered branch. If *folder* is omitted,
    defaults to the folder of the current branch. Errors if the current branch
    is not the max branch in the folder. If *folder* is specified, creates
    from the max branch regardless of current branch.

**squash**
:   When called with the max branch in a folder checked out, creates the next
    numbered branch and squashes all commits since the divergence from the
    trunk branch (main/master) into one commit. Detects the trunk branch
    automatically. Errors if the current branch is not a folder branch or not
    the max branch.

**squash**
:   When called with the highest numbered branch in a folder checked out,
    creates the next numbered branch and squashes all commits since the
    divergence from the trunk branch (main/master) into one commit. Detects
    the trunk branch automatically. Errors if the current branch is not a
    folder branch or not the highest numbered branch.

**delete** *\<folder\>*
:   Delete all branches in the folder. Prompts for confirmation unless
    `--force` is set.

**delete-upto** *\<folder\>* *\<n\>*
:   Delete all numeric branches in the folder with suffix less than *n*.
    Non-numeric branches (e.g. `async/temp`) are preserved. Prompts for
    confirmation unless `--force` is set.

**rename** *\<old\>* *\<new\>*
:   Rename all branches from prefix *old* to *new*. Prompts for confirmation
    unless `--force` is set.

# LOW-LEVEL COMMANDS (PLUMBING)

**max branch** *\<folder\>*
:   Print the full name of the max branch in the folder (e.g. `async/4`).

**max number** *\<folder\>*
:   Print the numeric suffix of the max branch in the folder (e.g. `4`).

# OPTIONS

**--force**, **-f**
:   Skip confirmation prompts. For **delete**, also allows deleting the
    currently checked-out branch by detaching HEAD first.

# NOTES

When *[folder]* is omitted, the folder is inferred from the current branch
name by taking the substring before the first `/`. If the current branch
contains no `/`, the command fails.

Numeric branch suffixes may be integers or decimals (e.g. `2`, `2.5`).
Non-numeric suffixes (e.g. `temp`, `bigbooty`) are treated as opaque and
ignored by **increment** and **delete-upto**.

# COMPLETION

zsh completion is available via:

```
git folder completion > ~/.zsh/completions/_git-folder
```

# EXAMPLES

Start a numbered branch sequence:

```
$ git checkout -b async/1
$ # ... make 7 commits ...
$ git folder increment
creating async/2
$ git rebase -i main
```

List branches in a folder:

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

Delete old branches, keeping from 4 onward:

```
$ git folder delete-upto async 4
keep:
  async/4
  async/bigbooty
  async/temp

delete:
  async/1
  async/2
  async/2.5
  async/3
confirm? [yN]
```

Clean up after a PR merges:

```
$ git folder delete async
```

# SEE ALSO

**git**(1), **git-branch**(1)
