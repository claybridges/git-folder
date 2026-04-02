package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/claybridges/git-folder/internal/folder"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "list", "ls":
		err = cmdList(os.Args[2:])
	case "increment", "inc":
		err = cmdIncrement(os.Args[2:])
	case "last-number", "ln":
		err = cmdLastNumber(os.Args[2:])
	case "delete", "rm":
		err = cmdDelete(os.Args[2:])
	case "rename", "mv":
		err = cmdRename(os.Args[2:])
	case "completion":
		cmdCompletion()
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage: git folder <command> [args]

commands:
  list        (ls) <folder>          list branches in folder
  last-number (ln) <folder>         print highest numbered branch
  increment   (inc) [folder]        create next numbered branch
  delete      (rm) <folder>         delete all branches in folder
  rename      (mv) <old> <new>      rename folder prefix
`)
}

func cmdList(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git folder list <folder>")
	}

	branches, err := folder.Enumerate(args[0])
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		return fmt.Errorf("no branches in folder %s/", args[0])
	}

	for _, b := range branches {
		fmt.Println(b)
	}
	return nil
}

func cmdLastNumber(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git folder last-number <folder>")
	}

	n, err := folder.LastNumber(args[0])
	if err != nil {
		return err
	}

	fmt.Println(n)
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
	} else {
		return fmt.Errorf("usage: git folder increment [folder]")
	}

	last, err := folder.LastNumber(name)
	if err != nil {
		return err
	}

	next := fmt.Sprintf("%s/%d", name, last+1)
	fmt.Printf("creating %s\n", next)
	return gitExec("checkout", "-b", next)
}

func cmdDelete(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git folder delete <folder>")
	}

	branches, err := folder.Enumerate(args[0])
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		return fmt.Errorf("no branches in folder %s/", args[0])
	}

	fmt.Printf("delete all branches in folder %s/:\n", args[0])
	for _, b := range branches {
		fmt.Printf("  %s\n", b)
	}

	fmt.Print("confirm? y/N ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" {
		fmt.Println("aborted")
		return nil
	}

	for _, b := range branches {
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

	// Check for conflicts
	existing, _ := folder.Enumerate(newName)
	if len(existing) > 0 {
		return fmt.Errorf("target folder %s/ already has branches", newName)
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

	fmt.Print("confirm? y/N ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" {
		fmt.Println("aborted")
		return nil
	}

	for _, p := range pairs {
		if err := gitExec("branch", "-m", p.from, p.to); err != nil {
			return fmt.Errorf("failed to rename %s -> %s: %w", p.from, p.to, err)
		}
	}
	return nil
}

func cmdCompletion() {
	fmt.Print(`#compdef git-folder

# Place in your fpath or run:
#   git-folder completion > ~/.zsh/completions/_git-folder

_git-folder() {
    local -a commands
    commands=(
        'list:list branches in a folder'
        'ls:list branches in a folder'
        'last-number:print highest numbered branch'
        'ln:print highest numbered branch'
        'increment:create next numbered branch'
        'inc:create next numbered branch'
        'delete:delete all branches in a folder'
        'rm:delete all branches in a folder'
        'rename:rename a folder prefix'
        'mv:rename a folder prefix'
        'completion:output zsh completion script'
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
                list|ls|last-number|ln|delete|rm|increment|inc|rename|mv)
                    _git-folder-folders
                    ;;
            esac
            ;;
    esac
}

_git-folder-folders() {
    local -a folders
    folders=(${(u)${(f)"$(git branch --format='%(refname:short)' 2>/dev/null | grep '/' | sed 's|/.*||')"}})
    [[ ${#folders} -gt 0 ]] && compadd -- "${folders[@]}"
}

# Also hook into 'git folder' as a git subcommand
_git-folder "$@"
`)
}

func gitExec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
