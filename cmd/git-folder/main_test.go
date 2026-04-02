package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	return dir
}

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func inDir(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func branchList(t *testing.T, dir, pattern string) []string {
	t.Helper()
	out := run(t, dir, "git", "branch", "--list", pattern, "--format=%(refname:short)")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.WriteString(input)
	w.Close()
	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = orig })
}

// --- cmdList ---

func TestCmdList(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/1")
	run(t, dir, "git", "branch", "foo/2")

	err := cmdList([]string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCmdListEmpty(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	err := cmdList([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for empty folder")
	}
}

func TestCmdListBadArgs(t *testing.T) {
	if err := cmdList(nil); err == nil {
		t.Fatal("expected error for no args")
	}
	if err := cmdList([]string{"a", "b"}); err == nil {
		t.Fatal("expected error for too many args")
	}
}

// --- cmdLastNumber ---

func TestCmdLastNumber(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/1")
	run(t, dir, "git", "branch", "foo/5")
	run(t, dir, "git", "branch", "foo/3")

	err := cmdLastNumber([]string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCmdLastNumberEmpty(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	err := cmdLastNumber([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent folder")
	}
}

func TestCmdLastNumberBadArgs(t *testing.T) {
	if err := cmdLastNumber(nil); err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestCmdLastNumberNotARepo(t *testing.T) {
	inNonRepo(t)
	if err := cmdLastNumber([]string{"foo"}); err == nil {
		t.Fatal("expected error outside git repo")
	}
}

// --- cmdIncrement ---

func TestCmdIncrementExplicit(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "topic/1")
	run(t, dir, "git", "branch", "topic/3")

	err := cmdIncrement([]string{"topic"})
	if err != nil {
		t.Fatal(err)
	}

	branches := branchList(t, dir, "topic/*")
	found := false
	for _, b := range branches {
		if b == "topic/4" {
			found = true
		}
	}
	if !found {
		t.Errorf("topic/4 not created, got branches: %v", branches)
	}
}

func TestCmdIncrementInferred(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "-b", "work/1")
	run(t, dir, "git", "branch", "work/2")

	err := cmdIncrement(nil)
	if err != nil {
		t.Fatal(err)
	}

	branches := branchList(t, dir, "work/*")
	found := false
	for _, b := range branches {
		if b == "work/3" {
			found = true
		}
	}
	if !found {
		t.Errorf("work/3 not created, got branches: %v", branches)
	}
}

func TestCmdIncrementNoFolder(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	// On main, not a folder branch
	err := cmdIncrement(nil)
	if err == nil {
		t.Fatal("expected error when not on folder branch")
	}
}

func TestCmdIncrementNonexistent(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	err := cmdIncrement([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent folder")
	}
}

func TestCmdIncrementBadArgs(t *testing.T) {
	if err := cmdIncrement([]string{"a", "b"}); err == nil {
		t.Fatal("expected error for too many args")
	}
}

// --- cmdDelete ---

func TestCmdDeleteConfirm(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "del/1")
	run(t, dir, "git", "branch", "del/2")

	withStdin(t, "y\n")

	err := cmdDelete([]string{"del"})
	if err != nil {
		t.Fatal(err)
	}

	branches := branchList(t, dir, "del/*")
	if len(branches) != 0 {
		t.Errorf("branches not deleted: %v", branches)
	}
}

func TestCmdDeleteAbort(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "keep/1")
	run(t, dir, "git", "branch", "keep/2")

	withStdin(t, "n\n")

	err := cmdDelete([]string{"keep"})
	if err != nil {
		t.Fatal(err)
	}

	branches := branchList(t, dir, "keep/*")
	if len(branches) != 2 {
		t.Errorf("branches should be kept, got: %v", branches)
	}
}

func TestCmdDeleteEmpty(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	err := cmdDelete([]string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for empty folder")
	}
}

func TestCmdDeleteBadArgs(t *testing.T) {
	if err := cmdDelete(nil); err == nil {
		t.Fatal("expected error for no args")
	}
}

// --- cmdRename ---

func TestCmdRenameConfirm(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "old/1")
	run(t, dir, "git", "branch", "old/2")

	withStdin(t, "y\n")

	err := cmdRename([]string{"old", "new"})
	if err != nil {
		t.Fatal(err)
	}

	oldBranches := branchList(t, dir, "old/*")
	newBranches := branchList(t, dir, "new/*")
	if len(oldBranches) != 0 {
		t.Errorf("old branches still exist: %v", oldBranches)
	}
	if len(newBranches) != 2 {
		t.Errorf("expected 2 new branches, got: %v", newBranches)
	}
}

func TestCmdRenameAbort(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "src/1")

	withStdin(t, "n\n")

	err := cmdRename([]string{"src", "dst"})
	if err != nil {
		t.Fatal(err)
	}

	if branches := branchList(t, dir, "src/*"); len(branches) != 1 {
		t.Errorf("source branches should remain: %v", branches)
	}
}

func TestCmdRenameConflict(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "aaa/1")
	run(t, dir, "git", "branch", "bbb/1")

	err := cmdRename([]string{"aaa", "bbb"})
	if err == nil {
		t.Fatal("expected error when target folder has branches")
	}
}

func TestCmdRenameEmpty(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	err := cmdRename([]string{"nonexistent", "whatever"})
	if err == nil {
		t.Fatal("expected error for empty source folder")
	}
}

func TestCmdRenameBadArgs(t *testing.T) {
	if err := cmdRename(nil); err == nil {
		t.Fatal("expected error for no args")
	}
	if err := cmdRename([]string{"a"}); err == nil {
		t.Fatal("expected error for one arg")
	}
}

// --- Not-a-repo error paths ---

func inNonRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	inDir(t, dir)
}

func TestCmdListNotARepo(t *testing.T) {
	inNonRepo(t)
	if err := cmdList([]string{"foo"}); err == nil {
		t.Fatal("expected error outside git repo")
	}
}

func TestCmdIncrementNotARepo(t *testing.T) {
	inNonRepo(t)
	if err := cmdIncrement([]string{"foo"}); err == nil {
		t.Fatal("expected error outside git repo")
	}
}

func TestCmdDeleteNotARepo(t *testing.T) {
	inNonRepo(t)
	if err := cmdDelete([]string{"foo"}); err == nil {
		t.Fatal("expected error outside git repo")
	}
}

func TestCmdRenameNotARepo(t *testing.T) {
	inNonRepo(t)
	if err := cmdRename([]string{"foo", "bar"}); err == nil {
		t.Fatal("expected error outside git repo")
	}
}

// --- main() / usage() via subprocess ---

func TestMainHelp(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperMain", "--", "help")
	cmd.Env = append(os.Environ(), "GIT_FOLDER_TEST_MAIN=1")
	out, _ := cmd.CombinedOutput()
	if !strings.Contains(string(out), "usage:") {
		t.Errorf("expected usage output, got: %s", out)
	}
}

func TestMainNoArgs(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperMain", "--")
	cmd.Env = append(os.Environ(), "GIT_FOLDER_TEST_MAIN=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for no args")
	}
}

func TestMainUnknownCommand(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperMain", "--", "bogus")
	cmd.Env = append(os.Environ(), "GIT_FOLDER_TEST_MAIN=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for unknown command")
	}
	if !strings.Contains(string(out), "unknown command") {
		t.Errorf("expected 'unknown command', got: %s", out)
	}
}

// Helper process that runs main() — invoked by TestMain* tests above.
func TestHelperMain(t *testing.T) {
	if os.Getenv("GIT_FOLDER_TEST_MAIN") != "1" {
		return
	}
	// Strip test flags, keep args after "--"
	args := []string{"git-folder"}
	for i, a := range os.Args {
		if a == "--" {
			args = append(args, os.Args[i+1:]...)
			break
		}
	}
	os.Args = args
	main()
}

// --- Delete failure (branch checked out) ---

func TestCmdDeleteFailsOnCheckedOut(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "-b", "doomed/1")

	withStdin(t, "y\n")

	err := cmdDelete([]string{"doomed"})
	if err == nil {
		t.Fatal("expected error deleting checked-out branch")
	}
}
