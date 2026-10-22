//usr/bin/env go run "$0" "$@"; exit
//go:build ignore

// Symlinks the committed hooks into the repo's shared hooks dir, which all worktrees use. The link points into the
// checkout this runs from, so edits to scripts/ take effect without re-installing.
//
// Usage: installHooks.go [--uninstall]
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	scriptsDir := scriptsDirectory()
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "not in a git repo:", err)
		os.Exit(1)
	}
	commonDir, _ := filepath.Abs(strings.TrimSpace(string(out)))
	hooksDir := filepath.Join(commonDir, "hooks")
	if out, err := exec.Command("git", "config", "--get", "core.hooksPath").Output(); err == nil {
		if path := strings.TrimSpace(string(out)); path != "" {
			hooksDir, _ = filepath.Abs(path)
		}
	}

	if len(os.Args) > 1 && os.Args[1] == "--uninstall" {
		uninstall(scriptsDir, hooksDir, commonDir)
		return
	}
	install(scriptsDir, hooksDir)
}

// ours reports whether a hook symlink target is one of our committed hooks, wherever the checkout is.
func ours(name, target string) bool {
	return strings.HasSuffix(target, "/scripts/hooks/"+name)
}

func install(scriptsDir, hooksDir string) {
	must(os.MkdirAll(hooksDir, 0o755))
	if strings.Contains(scriptsDir, "/.worktrees/") {
		fmt.Fprintln(os.Stderr, "warning: installing from a worktree; the link dangles if it is removed."+
			" Prefer running this from the main checkout.")
	}
	entries, err := os.ReadDir(filepath.Join(scriptsDir, "hooks"))
	must(err)
	for _, entry := range entries {
		name := entry.Name()
		source := filepath.Join(scriptsDir, "hooks", name)
		target := filepath.Join(hooksDir, name)
		info, err := os.Lstat(target)
		switch {
		case err != nil:
			must(os.Symlink(source, target))
			fmt.Printf("%s: linked %s -> %s\n", name, target, source)
		case info.Mode()&os.ModeSymlink == 0:
			fmt.Fprintf(os.Stderr, "%s: refusing to overwrite existing file %s\n", name, target)
		default:
			current, _ := os.Readlink(target)
			switch {
			case current == source:
				fmt.Printf("%s: already installed\n", name)
			case ours(name, current):
				must(os.Remove(target))
				must(os.Symlink(source, target))
				fmt.Printf("%s: updated %s -> %s (was %s)\n", name, target, source, current)
			default:
				fmt.Fprintf(os.Stderr, "%s: refusing to overwrite symlink %s -> %s\n", name, target, current)
			}
		}
	}
}

// uninstall removes only symlinks to our hooks, dangling or not; anything else is left alone.
func uninstall(scriptsDir, hooksDir, commonDir string) {
	removed := false
	entries, err := os.ReadDir(filepath.Join(scriptsDir, "hooks"))
	must(err)
	for _, entry := range entries {
		name := entry.Name()
		target := filepath.Join(hooksDir, name)
		info, err := os.Lstat(target)
		switch {
		case err != nil:
		case info.Mode()&os.ModeSymlink == 0:
			fmt.Fprintf(os.Stderr, "%s: not a symlink, leaving %s\n", name, target)
		default:
			current, _ := os.Readlink(target)
			if !ours(name, current) {
				fmt.Fprintf(os.Stderr, "%s: links to %s, not ours, leaving %s\n", name, current, target)
				continue
			}
			must(os.Remove(target))
			fmt.Printf("%s: removed %s\n", name, target)
			removed = true
		}
	}
	if !removed {
		fmt.Println("nothing to remove")
	}
	fmt.Printf("log and lock files left in %s (rebuildDevBinary.*)\n", commonDir)
}

// scriptsDirectory is the directory holding this file. `go run` builds in a temp dir, so os.Executable won't do.
func scriptsDirectory() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
