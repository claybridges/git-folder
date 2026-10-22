//usr/bin/env go run "$0" "$@"; exit
//go:build ignore

// Copies the committed hooks (and the script they launch) into the repo's shared hooks dir, which all worktrees use.
// Copies, not symlinks, so the hook keeps working on checkouts that don't contain scripts/. Re-run after changing them.
package main

import (
	"bytes"
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
	must(os.MkdirAll(hooksDir, 0o755))

	// The support script is ours to overwrite; hooks are not, if they differ from what we'd install.
	copyFile(filepath.Join(scriptsDir, "rebuildDevBinary.go"), filepath.Join(hooksDir, "rebuildDevBinary.go"))

	entries, err := os.ReadDir(filepath.Join(scriptsDir, "hooks"))
	must(err)
	for _, entry := range entries {
		name := entry.Name()
		source := filepath.Join(scriptsDir, "hooks", name)
		target := filepath.Join(hooksDir, name)
		want, err := os.ReadFile(source)
		must(err)
		have, err := os.ReadFile(target)
		switch {
		case err != nil:
			copyFile(source, target)
			fmt.Printf("%s: installed %s\n", name, target)
		case bytes.Equal(have, want):
			fmt.Printf("%s: already installed\n", name)
		case bytes.Contains(have, []byte("rebuildDevBinary.go")):
			copyFile(source, target)
			fmt.Printf("%s: updated %s\n", name, target)
		default:
			fmt.Fprintf(os.Stderr, "%s: refusing to overwrite different existing %s\n", name, target)
		}
	}
}

// scriptsDirectory is the directory holding this file. `go run` builds in a temp dir, so os.Executable won't do.
func scriptsDirectory() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func copyFile(source, target string) {
	data, err := os.ReadFile(source)
	must(err)
	tmp := target + ".tmp"
	must(os.WriteFile(tmp, data, 0o755))
	must(os.Rename(tmp, target))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
