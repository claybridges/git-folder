# include files
if [[ -n "${ZSH_VERSION:-}" ]]; then
    eval 'thisDir="${${(%):-%x}:h}"'
else
    thisDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi
source "${thisDir}/common.sh"

# Functions for handling branch folders

gitBranchFolderEnumerate() {
    local folderName="$1"
    git branch --format='%(refname:short)' | grep "^${folderName}/"
}

gitBranchFolderIsValid() {
    # Check for the format {text}/{text}
    [[ "$1" =~ ^[^/]+/[^/]+$ ]]
}

gitBranchFolderExists() {
    local branch="$1"

    if ! gitBranchFolderIsValid "$branch"; then
        echoError "${branch} is not a branch folder"
        return 1
    fi

    git show-ref --verify --quiet "refs/heads/$branch"
}

gitBranchFolderNumber() {
    local branch="$1"
    if ! gitBranchFolderIsValid "$branch"; then
        echoError "Invalid git branch folder '$branch'" || return 1
    fi

    local rhs="${branch#*/}"
    # Check if the right hand side is a positive integer
    if [[ "$rhs" =~ ^[0-9]+$ ]]; then
        echo "$rhs"
    fi
}

gitBranchFolderName() {
    local branch="$1"
    if ! gitBranchFolderIsValid "$branch"; then
        echo "Error: Invalid git branch folder '$branch'" >&2
        exit 1
    fi

    echo "${branch%%/*}"
}

gitBranchFolderCurrent() {
    local currentBranch=$(git symbolic-ref --short HEAD 2>/dev/null)

    # If a branch is checked out
    if [[ -n "$currentBranch" ]]; then
        if gitBranchFolderIsValid "$currentBranch"; then
            echo "$currentBranch"
            return
        fi
    fi

    # Handle detached HEAD: get branches coincident with current commit
    local currentSha=$(git rev-parse HEAD)
    # local branches=$(git branch --contains "$currentSha" --format="%(refname:short)")
    # local branches=$(git branch --points-at "$currentSha" --format="%(refname:short)")
    local branches=$(
       git for-each-ref \
           --points-at "$currentSha" \
           --format="%(refname:short)" \
           refs/heads
    )


    [ -z "$branches" ] && return

    # 3. Filter branches that pass isValid
    local validBranches=()
    while IFS= read -r b; do
        [[ -z "$b" ]] && continue
        if gitBranchFolderIsValid "$b"; then
            validBranches+=("$b")
        fi
    done <<< "$branches"

    if [[ ${#validBranches[@]} -eq 1 ]]; then
        echo "${validBranches[@]:0:1}"
    elif [[ ${#validBranches[@]} -gt 1 ]]; then
        local list=$(echo "${validBranches[@]}" | tr ' ' ',')
        local shortSha="$(git rev-parse --short HEAD | tr -d '\n')"
        echoError "Ambigous branches at ${shortSha}:\n    ${list}"
        (exit 1)
        # return 1
    fi
}

gitBranchFolderLastNumber() {
    local branchFolder="$1"
    local max=-1
    local found=false
    local num

    # Get all local branches starting with the folder prefix
    local branches
    branches=$(git branch --list "${branchFolder}/*" --format="%(refname:short)")

    while IFS= read -r b; do
        [[ -z "$b" ]] && continue
        num=$(gitBranchFolderNumber "$b")
        if [[ -n "$num" ]]; then
            found=true
            if [[ "$num" -gt "$max" ]]; then
                max=$num
            fi
        fi
    done <<< "$branches"

    if [ "$found" = true ]; then
        echo "$max"
    else
        echo "No branches like ${branchFolder}/<number>" >&2
        return 1
    fi
}

# These are the main aliased commands

gitBranchFolderDelete() {
    if [ $# -ne 1 ]; then
        echoBold "One argument required: folderName"
        return 1
    fi

    local branches=($(gitBranchFolderEnumerate "$1"))

    if [[ ${#branches[@]} -eq 0 ]]; then
        echoError "No branches in folder $1/"
        return 1
    fi

    echoBold "Delete all branches in folder $1/:"
    for b in "${branches[@]}"; do
        echo "    $b"
    done

    printf "Confirm deletion of all above branches? yN "; read -r confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || { echoBold "Rejected"; return 1; }

    for b in "${branches[@]}"; do
        git branch -D "$b"
    done
}

gitBranchFolderIncrement() {
    local folder

    if [[ $# -eq 1 ]]; then
        # Explicit folder name: find max and increment
        folder="$1"
        local lastNumber
        lastNumber=$(gitBranchFolderLastNumber "$folder") || return 1
        local nextNumber=$((lastNumber + 1))
        git checkout -b "${folder}/${nextNumber}"
        return $?
    fi

    # No argument: infer from current branch
    local branch
    branch=$(gitBranchFolderCurrent)

    if [[ $? -ne 0 ]]; then
        return 1
    fi

    if [[ -z "$branch" ]]; then
        echoError "No current branch folder" && return 1
    fi

    folder=$(gitBranchFolderName "$branch")
    local currentNumber=$(gitBranchFolderNumber "$branch")

    if [[ -z "$currentNumber" ]]; then
        echoError "Current branch folder is not numbered: $branch"
        return 1
    fi

    local lastNumber
    lastNumber=$(gitBranchFolderLastNumber "$folder") || return 1

    if (( lastNumber != currentNumber )); then
        echoError "Current folder branch $branch is not max $folder/$lastNumber"
        return 1
    fi

    local nextNumber=$((lastNumber + 1))
    git checkout -b "${folder}/${nextNumber}"
}

gitBranchFolderList() {
    if [ $# -ne 1 ]; then
        echoBold "One argument required: folderName"
        return 1
    fi

    local branches=($(gitBranchFolderEnumerate "$1"))

    if [[ ${#branches[@]} -eq 0 ]]; then
        echoError "No branches in folder $1/"
        return 1
    fi

    for b in "${branches[@]}"; do
        echo "$b"
    done
}

# renames branch folder, e.g.  branches `foo/{1,2,3}` -> `bar/{1,2,3}`
# - lots of safety checks, and confirmation prompt
gitBranchFolderRename() {
    if [ $# -ne 2 ]; then
        echoBold "Two arguments required: sourceFolderName targetFolderName"
        return 1
    fi

    local sourceFolderName="$1"
    local targetFolderName="$2"

    local sources=($(git branch --format='%(refname:short)' | grep "^${sourceFolderName}/"))

    if [[ ${#sources[@]} -eq 0 ]]; then
        echoError "No branches in folder ${sourceFolderName}/"
        return 1
    fi

    # make sure no target branches exist
    local conflicts=()
    for b in "${sources[@]}"; do
        local target="${b/#${sourceFolderName}\//${targetFolderName}/}"
        if git show-ref --verify --quiet "refs/heads/$target"; then
            conflicts+=("$target")
        fi
    done

    if [[ ${#conflicts[@]} -gt 0 ]]; then
        echoError "Target branch(es) exist, exiting:"
        for c in "${conflicts[@]}"; do
            echoError "    $c"
        done
        return 1
    fi

    echoBold "Rename branch folder ${sourceFolderName}/ -> ${targetFolderName}/:"
    for b in "${sources[@]}"; do
        echo "    $b -> ${b/${sourceFolderName}\//${targetFolderName}/}"
    done

    printf "Confirm: yN "; read -r confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || { echoBold "Rejected"; return 1; }

    for b in "${sources[@]}"; do
        local target="${b/#${sourceFolderName}\//${targetFolderName}/}"
        git branch -m "$b" "$target"
    done
}
