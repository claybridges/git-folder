package main

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/claybridges/git-folder/internal/folder"
)

//go:embed USAGE.txt
var usageText string

var version = "dev"
var forceFlag bool

func parseGlobalFlags(args []string) []string {
	var filtered []string
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			forceFlag = true
		} else {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

func confirm(prompt string) bool {
	if forceFlag {
		return true
	}
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y"
}

// branchesPreflight validates that the given branches can be modified.
// It returns the name of the current branch if it needs to be detached
// (i.e. --force is set and the current branch is in the list), or "" if not.
// It fails closed: errors from git symbolic-ref and git worktree list are returned.
func branchesPreflight(branches []string) (detachBranch string, err error) {
	// Get current branch; exit code 1 = detached HEAD (ok), anything else = error.
	out, symErr := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD").Output()
	current := strings.TrimSpace(string(out))
	if symErr != nil {
		var exitErr *exec.ExitError
		if errors.As(symErr, &exitErr) && exitErr.ExitCode() == 1 {
			current = "" // detached HEAD
		} else {
			return "", fmt.Errorf("failed to determine current branch: %w", symErr)
		}
	}

	// Check worktrees — fail closed if this command fails.
	wtOut, wtErr := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if wtErr != nil {
		return "", fmt.Errorf("failed to check git worktrees: %w", wtErr)
	}
	var inWorktrees []string
	for _, line := range strings.Split(string(wtOut), "\n") {
		if !strings.HasPrefix(line, "branch ") {
			continue
		}
		wtBranch := strings.TrimPrefix(line, "branch refs/heads/")
		if wtBranch == current {
			continue // handled below
		}
		for _, b := range branches {
			if b == wtBranch {
				inWorktrees = append(inWorktrees, b)
			}
		}
	}
	if len(inWorktrees) == 1 {
		return "", fmt.Errorf("cannot modify branch '%s': checked out in a worktree", inWorktrees[0])
	} else if len(inWorktrees) > 1 {
		return "", fmt.Errorf("cannot modify branches: checked out in worktrees: %s", strings.Join(inWorktrees, ", "))
	}

	// Check if current branch is in list.
	for _, b := range branches {
		if b == current {
			if forceFlag {
				return current, nil // caller will detach before first mutation
			}
			return "", fmt.Errorf("cannot modify branch '%s': currently checked out (use --force to detach)", b)
		}
	}
	return "", nil
}

// detachHEAD detaches HEAD from the named branch, printing a message.
func detachHEAD(branch string) error {
	fmt.Printf("Detaching HEAD from '%s' (branch is being modified)\n", branch)
	if err := gitExec("checkout", "--detach"); err != nil {
		return fmt.Errorf("failed to detach HEAD: %w", err)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	// Parse global flags
	args := parseGlobalFlags(os.Args[1:])
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "list":
		err = cmdList(args[1:])
	case "increment":
		err = cmdIncrement(args[1:])
	case "max":
		err = cmdMax(args[1:])
	case "delete":
		err = cmdDelete(args[1:])
	case "delete-upto":
		err = cmdDeleteUpto(args[1:])
	case "squash":
		err = cmdSquash()
	case "rename":
		err = cmdRename(args[1:])
	case "completion":
		cmdCompletion()
	case "version", "--version", "-v":
		fmt.Println(version)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, usageText)
}

func resolveFolder(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) == 0 {
		cur, err := folder.CurrentFolder()
		if err != nil {
			return "", err
		}
		if cur == "" {
			return "", fmt.Errorf("not on a folder branch; pass folder name explicitly")
		}
		return cur, nil
	}
	return "", fmt.Errorf("too many arguments")
}

func cmdList(args []string) error {
	name, err := resolveFolder(args)
	if err != nil {
		return err
	}

	branches, err := folder.Enumerate(name)
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		return fmt.Errorf("no branches in folder %s/", name)
	}

	for _, b := range branches {
		fmt.Println(b)
	}
	return nil
}

func cmdMax(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: git folder max branch|number <folder>")
	}
	sub, folderName := args[0], args[1]

	switch sub {
	case "branch":
		b, err := folder.MaxBranch(folderName)
		if err != nil {
			return err
		}
		fmt.Println(b)
	case "number":
		n, err := folder.LastNumber(folderName)
		if err != nil {
			return err
		}
		if n == float64(int(n)) {
			fmt.Println(int(n))
		} else {
			fmt.Println(n)
		}
	default:
		return fmt.Errorf("unknown max subcommand %q; use branch or number", sub)
	}
	return nil
}

func cmdIncrement(args []string) error {
	var name string

	if len(args) == 1 {
		name = args[0]
	} else if len(args) == 0 {
		cur, err := folder.CurrentFolder()
		if err != nil {
			return err
		}
		if cur == "" {
			return fmt.Errorf("not on a folder branch; pass folder name explicitly")
		}
		name = cur

		// When inferring, verify we're on the max branch
		cur, err2 := folder.CurrentBranch()
		if err2 != nil {
			return err2
		}
		max, err2 := folder.MaxBranch(name)
		if err2 != nil {
			return err2
		}
		if cur != max {
			return fmt.Errorf("current branch is %s but max branch is %s", cur, max)
		}
	} else {
		return fmt.Errorf("usage: git folder increment [folder]")
	}

	last, err := folder.LastNumber(name)
	if err != nil {
		return err
	}

	ceiled := int(math.Ceil(last))
	next := ceiled
	if float64(ceiled) == last {
		next = ceiled + 1
	}
	nextBranch := fmt.Sprintf("%s/%d", name, next)
	fmt.Printf("creating %s\n", nextBranch)
	return gitExec("checkout", "-b", nextBranch)
}

func cmdDelete(args []string) error {
	name, err := resolveFolder(args)
	if err != nil {
		return err
	}

	branches, err := folder.Enumerate(name)
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		return fmt.Errorf("no branches in folder %s/", name)
	}

	detach, err := branchesPreflight(branches)
	if err != nil {
		return err
	}

	fmt.Printf("delete all branches in folder %s/:\n", name)
	for _, b := range branches {
		fmt.Printf("  %s\n", b)
	}

	if !confirm("confirm? [yN] ") {
		fmt.Println("aborted")
		return nil
	}

	if detach != "" {
		if err := detachHEAD(detach); err != nil {
			return err
		}
	}
	for _, b := range branches {
		if err := gitExec("branch", "-D", b); err != nil {
			return fmt.Errorf("failed to delete %s: %w", b, err)
		}
	}
	return nil
}

func cmdDeleteUpto(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: git folder delete-upto <folder> <n>")
	}

	folderName := args[0]
	n, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid number: %s", args[1])
	}

	branches, err := folder.Enumerate(folderName)
	if err != nil {
		return err
	}

	var toKeep []string
	var toDelete []string
	for _, b := range branches {
		num, ok := folder.NumberFloat(b)
		if ok && num < n {
			toDelete = append(toDelete, b)
		} else {
			toKeep = append(toKeep, b)
		}
	}

	if len(toDelete) == 0 {
		return fmt.Errorf("no numbered branches below %v in folder %s/", n, folderName)
	}

	detach, err := branchesPreflight(toDelete)
	if err != nil {
		return err
	}

	fmt.Printf("keep:\n")
	for _, b := range toKeep {
		fmt.Printf("  %s\n", b)
	}
	fmt.Println()
	fmt.Printf("delete:\n")
	for _, b := range toDelete {
		fmt.Printf("  %s\n", b)
	}

	if !confirm("confirm? [yN] ") {
		fmt.Println("aborted")
		return nil
	}

	if detach != "" {
		if err := detachHEAD(detach); err != nil {
			return err
		}
	}
	for _, b := range toDelete {
		if err := gitExec("branch", "-D", b); err != nil {
			return fmt.Errorf("failed to delete %s: %w", b, err)
		}
	}
	return nil
}

func cmdRename(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: git folder rename <old> <new>")
	}

	oldName, newName := args[0], args[1]

	sources, err := folder.Enumerate(oldName)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no branches in folder %s/", oldName)
	}

	// Check for conflicts before any side effects
	existing, err := folder.Enumerate(newName)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return fmt.Errorf("target folder %s/ already has branches", newName)
	}

	detach, err := branchesPreflight(sources)
	if err != nil {
		return err
	}

	// Build rename pairs
	type pair struct{ from, to string }
	pairs := make([]pair, len(sources))
	for i, src := range sources {
		suffix := src[len(oldName):]
		pairs[i] = pair{src, newName + suffix}
	}

	fmt.Printf("rename %s/ -> %s/:\n", oldName, newName)
	for _, p := range pairs {
		fmt.Printf("  %s -> %s\n", p.from, p.to)
	}

	if !confirm("confirm? [yN] ") {
		fmt.Println("aborted")
		return nil
	}

	if detach != "" {
		if err := detachHEAD(detach); err != nil {
			return err
		}
	}
	for _, p := range pairs {
		if err := gitExec("branch", "-m", p.from, p.to); err != nil {
			return fmt.Errorf("failed to rename %s -> %s: %w", p.from, p.to, err)
		}
	}
	return nil
}

func cmdSquash() error {
	// Bail if working tree or index is dirty
	if err := exec.Command("git", "diff", "--quiet").Run(); err != nil {
		return fmt.Errorf("working tree has uncommitted changes; commit or stash first")
	}
	if err := exec.Command("git", "diff", "--cached", "--quiet").Run(); err != nil {
		return fmt.Errorf("index has staged changes; commit or stash first")
	}

	// Detect trunk
	trunk, err := folder.DetectTrunk()
	if err != nil {
		return err
	}

	// Find merge base with trunk
	mergeBase, err := gitOutput("merge-base", "HEAD", trunk)
	if err != nil {
		return fmt.Errorf("no common ancestor with %s: %w", trunk, err)
	}

	// Find first commit after split (topo-order, reversed for deterministic result)
	revList, err := gitOutput("rev-list", "--reverse", "--topo-order", "--ancestry-path", mergeBase+"..HEAD")
	if err != nil || revList == "" {
		return fmt.Errorf("no commits after divergence from %s", trunk)
	}
	lines := strings.Split(revList, "\n")
	firstCommit := lines[0]

	// Count commits to squash (everything after firstCommit)
	countStr, err := gitOutput("rev-list", "--count", firstCommit+"..HEAD")
	if err != nil {
		return err
	}
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return fmt.Errorf("unexpected output from git rev-list --count: %q", countStr)
	}
	if count == 0 {
		return fmt.Errorf("only one commit on branch, nothing to squash")
	}

	// Increment (infers folder, validates we're on highest, creates next branch)
	if err = cmdIncrement(nil); err != nil {
		return err
	}

	// Collect messages from commits being squashed
	msgs, err := gitOutput("log", "--format=%B", firstCommit+"..HEAD")
	if err != nil {
		return err
	}
	var filtered []string
	for _, line := range strings.Split(msgs, "\n") {
		if strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}

	// Get first commit's message
	firstMsg, err := gitOutput("log", "-1", "--format=%B", firstCommit)
	if err != nil {
		return err
	}

	// Squash: reset to first commit, amend with all messages
	if err = gitExec("reset", "--soft", firstCommit); err != nil {
		return err
	}

	combinedMsg := strings.TrimSpace(firstMsg) + "\n\n" + strings.Join(filtered, "\n")
	return gitExec("commit", "--amend", "-m", combinedMsg)
}

func cmdCompletion() {
	fmt.Print(`#compdef git-folder

# Place in your fpath or run:
#   git-folder completion > ~/.zsh/completions/_git-folder

_git-folder() {
    local -a commands
    commands=(
        'list:list branches in a folder'
        'max:print max branch or number'
        'increment:create next numbered branch'
        'delete:delete all branches in a folder'
        'delete-upto:delete numbered branches below n'
        'squash:increment and squash commits'
        'rename:rename a folder prefix'
        'version:show version'
        'help:show usage'
    )

    _arguments -C \
        '1:command:->cmd' \
        '*::arg:->args'

    case $state in
        cmd)
            _describe 'command' commands
            ;;
        args)
            case $words[1] in
                list|delete|delete-upto|increment|rename)
                    _git-folder-folders
                    ;;
            esac
            ;;
    esac
}

_git-folder-folders() {
    local -a folders
    folders=(${(u)${(f)"$(
        git branch --format='%(refname:short)' 2>/dev/null \
            | grep -E '^[^/]+/[^/]+$' \
            | sed 's|/.*||'
    )"}})
    [[ ${#folders} -gt 0 ]] && compadd -- "${folders[@]}"
}

# Also hook into 'git folder' as a git subcommand
_git-folder "$@"
`)
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func gitExec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
