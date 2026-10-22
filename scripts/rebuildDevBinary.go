//usr/bin/env go run "$0" "$@"; exit
//go:build ignore

// Builds main's tree and refreshes the local dev binary, if there is one. Run detached by hooks/reference-transaction.
//
// Usage: rebuildDevBinary.go [newOid]
// Env:   gitFolderBinDir       install dir (default: same resolution as the Taskfile's LOCAL_BIN_DIR)
//
//	gitFolderSkipInstall  build but don't install
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const binaryName = "git-folder"

var (
	logPath     string
	pendingPath string
)

func main() {
	if len(os.Args) > 1 && regexp.MustCompile(`^0+$`).MatchString(os.Args[1]) {
		return // branch deleted
	}
	for _, name := range []string{"GIT_INDEX_FILE", "GIT_PREFIX", "GIT_WORK_TREE"} {
		os.Unsetenv(name)
	}
	trigger := ""
	if len(os.Args) > 1 {
		trigger = short(os.Args[1])
	}

	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "not in a git repo:", err)
		os.Exit(1)
	}
	commonDir, _ := filepath.Abs(strings.TrimSpace(string(out)))
	logPath = filepath.Join(commonDir, "rebuildDevBinary.log")
	pendingPath = filepath.Join(commonDir, "rebuildDevBinary.pending")

	lock, err := os.OpenFile(filepath.Join(commonDir, "rebuildDevBinary.lock"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		logf("FAILED: %v", err)
		os.Exit(1)
	}
	defer lock.Close()

	for {
		os.WriteFile(pendingPath, nil, 0o644)
		// A build is running; it will see the pending mark and rebuild once more.
		if syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
			return
		}
		for {
			os.Remove(pendingPath)
			logf("building (trigger %s)", trigger)
			if err := build(); err != nil {
				logf("FAILED: %v", err)
			}
			if !exists(pendingPath) {
				break
			}
		}
		syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		// A trigger may have queued between the last check and the unlock.
		if !exists(pendingPath) {
			return
		}
	}
}

func build() error {
	oid, err := run("", "git", "rev-parse", "--verify", "refs/heads/main")
	if err != nil {
		return err
	}
	oid = strings.TrimSpace(oid)
	// Only refresh a dev binary that exists: use-dev creates it, use-brew deletes it, and while it's gone this hook
	// must not bring it back to shadow brew's copy.
	if dest := filepath.Join(binDir(), binaryName); os.Getenv("gitFolderSkipInstall") == "" && !exists(dest) {
		logf("skipped %s: no dev binary at %s (go tool task use-dev creates one)", short(oid), dest)
		return nil
	}
	tmp, err := os.MkdirTemp("", "gitFolderDev")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	src := filepath.Join(tmp, "src")
	if err := os.Mkdir(src, 0o755); err != nil {
		return err
	}
	if _, err := run("", "sh", "-c", `git archive "$1" | tar -x -C "$2"`, "sh", oid, src); err != nil {
		return err
	}
	built := filepath.Join(tmp, binaryName)
	if _, err := run(src, "go", "build", "-o", built, "./cmd/git-folder"); err != nil {
		return err
	}
	if os.Getenv("gitFolderSkipInstall") != "" {
		logf("built %s, install skipped", short(oid))
		return nil
	}
	dir := binDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(built)
	if err != nil {
		return err
	}
	staged := filepath.Join(dir, fmt.Sprintf(".%s.%d", binaryName, os.Getpid()))
	if err := os.WriteFile(staged, data, 0o755); err != nil {
		return err
	}
	if err := os.Rename(staged, filepath.Join(dir, binaryName)); err != nil {
		os.Remove(staged)
		return err
	}
	logf("installed %s to %s", short(oid), dir)
	return nil
}

func binDir() string {
	if dir := os.Getenv("gitFolderBinDir"); dir != "" {
		return dir
	}
	if dir := os.Getenv("XDG_BIN_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	local := filepath.Join(home, ".local", "bin")
	if info, err := os.Stat(local); err == nil && info.IsDir() {
		return local
	}
	return filepath.Join(home, "bin")
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out), nil
}

func logf(format string, args ...any) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
}

func short(oid string) string {
	if len(oid) > 12 {
		return oid[:12]
	}
	return oid
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
