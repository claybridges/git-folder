#!/usr/bin/env bash

# This is a stripped down version of my dotfiles git aliases
# Kept only the git branch folder related functions and aliases

set -u

# include files
if [[ -n "${ZSH_VERSION:-}" ]]; then
    eval 'thisDir="${${(%):-%x}:h}"'
else
    thisDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi
source "${thisDir}/common.sh"
source "${thisDir}/gitBranchFolder.sh"

# Terminal formatting
bold=`tput bold`
reset=`tput sgr0`

echoBold() {
    local msg="$1"
    echo -e "${bold}$msg${reset}"
}

# Git branch folder aliases
alias gbfd=gitBranchFolderDelete
alias gbfi=gitBranchFolderIncrement
alias gbfl=gitBranchFolderList
alias gbfr=gitBranchFolderRename
