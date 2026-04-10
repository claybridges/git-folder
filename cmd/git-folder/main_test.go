package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/claybridges/git-folder/internal/example"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "git-folder-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp) //nolint:errcheck // best-effort cleanup

	binaryPath = filepath.Join(tmp, "git-folder")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	os.Exit(m.Run())
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := example.InitRepo(dir); err != nil {
		t.Fatal(err)
	}
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
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	})
}

func withFolder(folder string, nums ...string) []string {
	branches := make([]string, len(nums))
	for i, n := range nums {
		branches[i] = folder + "/" + n
	}
	return branches
}

func assertBranchesExist(t *testing.T, dir string, branches []string) {
	t.Helper()
	have := branchList(t, dir, "*/*")
	for _, b := range branches {
		if !slices.Contains(have, b) {
			t.Fatalf("branch %s missing", b)
		}
	}
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
	if _, err := w.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		if err := r.Close(); err != nil {
			t.Errorf("failed to close pipe: %v", err)
		}
	})
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
	inNonRepo(t)
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

func TestCmdIncrementExplicitFloat(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "topic/1")
	run(t, dir, "git", "branch", "topic/2.5")

	err := cmdIncrement([]string{"topic"})
	if err != nil {
		t.Fatal(err)
	}

	branches := branchList(t, dir, "topic/*")
	found := false
	for _, b := range branches {
		if b == "topic/3" {
			found = true
		}
	}
	if !found {
		t.Errorf("topic/3 not created, got branches: %v", branches)
	}
}

func TestCmdIncrementInferred(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "-b", "work/2")
	run(t, dir, "git", "branch", "work/1")

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

func TestCmdIncrementNotMax(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "-b", "work/1")
	run(t, dir, "git", "branch", "work/3")

	// On work/1 but max is work/3, should error
	err := cmdIncrement(nil)
	if err == nil {
		t.Fatal("expected error when not on max branch")
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
	inNonRepo(t)
	if err := cmdDelete(nil); err == nil {
		t.Fatal("expected error for no args")
	}
}

// initAsyncRepo creates the async/ folder from the README example with realistic history.
func initAsyncRepo(t *testing.T) string {
	t.Helper()
	dir := initTestRepo(t)
	if err := example.Setup(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

// --- cmdDeleteUpto ---

func TestCmdDeleteUptoReadmeExample(t *testing.T) {
	dir := initAsyncRepo(t)

	deleted := withFolder("async", "1", "2", "2.5", "3")
	assertBranchesExist(t, dir, deleted)

	cmd := exec.Command(binaryPath, "delete-upto", "async", "4")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("delete-upto failed: %v\n%s", err, out)
	}

	kept := branchList(t, dir, "async/*")
	for _, b := range deleted {
		if slices.Contains(kept, b) {
			t.Errorf("branch %s should have been deleted", b)
		}
	}
}

func TestCmdDeleteUptoConfirm(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "x/1")
	run(t, dir, "git", "branch", "x/2")
	run(t, dir, "git", "branch", "x/2.5")
	run(t, dir, "git", "branch", "x/3")
	run(t, dir, "git", "branch", "x/4")
	run(t, dir, "git", "branch", "x/bigbooty")

	withStdin(t, "y\n")

	err := cmdDeleteUpto([]string{"x", "3"})
	if err != nil {
		t.Fatal(err)
	}

	kept := branchList(t, dir, "x/*")
	// Should keep: x/3, x/4, x/bigbooty
	if len(kept) != 3 {
		t.Errorf("expected 3 kept, got: %v", kept)
	}
	// Should have deleted: x/1, x/2, x/2.5
	for _, b := range kept {
		if b == "x/1" || b == "x/2" || b == "x/2.5" {
			t.Errorf("branch %s should have been deleted", b)
		}
	}
}

func TestCmdDeleteUptoFloatThreshold(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "x/1")
	run(t, dir, "git", "branch", "x/2")
	run(t, dir, "git", "branch", "x/2.5")
	run(t, dir, "git", "branch", "x/2.6")
	run(t, dir, "git", "branch", "x/3")

	withStdin(t, "y\n")

	err := cmdDeleteUpto([]string{"x", "2.5"})
	if err != nil {
		t.Fatal(err)
	}

	kept := branchList(t, dir, "x/*")
	// Should keep: x/2.5, x/2.6, x/3
	if len(kept) != 3 {
		t.Errorf("expected 3 kept, got: %v", kept)
	}
	for _, b := range kept {
		if b == "x/1" || b == "x/2" {
			t.Errorf("branch %s should have been deleted", b)
		}
	}
}

func TestCmdDeleteUptoAbort(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "y/1")
	run(t, dir, "git", "branch", "y/2")

	withStdin(t, "n\n")

	err := cmdDeleteUpto([]string{"y", "2"})
	if err != nil {
		t.Fatal(err)
	}

	kept := branchList(t, dir, "y/*")
	if len(kept) != 2 {
		t.Errorf("expected 2 kept, got: %v", kept)
	}
}

func TestCmdDeleteUptoNoneBelow(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "z/5")

	err := cmdDeleteUpto([]string{"z", "1"})
	if err == nil {
		t.Fatal("expected error when no branches below n")
	}
}

func TestCmdDeleteUptoBadArgs(t *testing.T) {
	if err := cmdDeleteUpto(nil); err == nil {
		t.Fatal("expected error for no args")
	}
	if err := cmdDeleteUpto([]string{"a"}); err == nil {
		t.Fatal("expected error for one arg")
	}
	if err := cmdDeleteUpto([]string{"a", "notanumber"}); err == nil {
		t.Fatal("expected error for non-numeric arg")
	}
}

func TestCmdDeleteUptoNotARepo(t *testing.T) {
	inNonRepo(t)
	if err := cmdDeleteUpto([]string{"foo", "3"}); err == nil {
		t.Fatal("expected error outside git repo")
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

// --- cmdSquash ---

func TestCmdSquash(t *testing.T) {
	dir := initTestRepo(t) // main with a.go, b.go, c.go

	// Create async/1 with 7 commits
	run(t, dir, "git", "checkout", "-b", "async/1")
	for i := 1; i <= 7; i++ {
		msg := fmt.Sprintf("async1-work-%d", i)
		if err := os.WriteFile(filepath.Join(dir, msg+".txt"), []byte(msg), 0644); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "git", "add", msg+".txt")
		run(t, dir, "git", "commit", "-m", msg)
	}

	// Run squash via binary
	cmd := exec.Command(binaryPath, "squash")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("squash failed: %v\n%s", err, out)
	}

	// Should be on async/2
	branch := run(t, dir, "git", "branch", "--show-current")
	if branch != "async/2" {
		t.Fatalf("expected async/2, got %s", branch)
	}

	// Should have exactly 1 commit between merge-base and HEAD
	mergeBase := run(t, dir, "git", "merge-base", "HEAD", "main")
	countStr := run(t, dir, "git", "rev-list", "--count", mergeBase+"..HEAD")
	if countStr != "1" {
		t.Fatalf("expected 1 commit after squash, got %s", countStr)
	}

	// Commit message should contain the squashed messages
	msg := run(t, dir, "git", "log", "-1", "--format=%B")
	for i := 1; i <= 7; i++ {
		want := fmt.Sprintf("async1-work-%d", i)
		if !strings.Contains(msg, want) {
			t.Errorf("commit message missing %q", want)
		}
	}
}

func TestCmdSquashNothingToSquash(t *testing.T) {
	dir := initTestRepo(t)

	// Create async/1 with only 1 commit
	run(t, dir, "git", "checkout", "-b", "async/1")
	if err := os.WriteFile(filepath.Join(dir, "only.txt"), []byte("only"), 0644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "only.txt")
	run(t, dir, "git", "commit", "-m", "only commit")

	cmd := exec.Command(binaryPath, "squash")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for nothing to squash")
	}
	if !strings.Contains(string(out), "nothing to squash") {
		t.Fatalf("expected 'nothing to squash' error, got: %s", out)
	}

	// async/2 should NOT exist (didn't increment)
	branches := branchList(t, dir, "async/*")
	for _, b := range branches {
		if b == "async/2" {
			t.Fatal("async/2 should not exist when nothing to squash")
		}
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

// --- Force flag tests ---

//gocyclo:ignore
func TestForceFlag(t *testing.T) {
	t.Run("delete with --force", func(t *testing.T) {
		dir := initTestRepo(t)
		inDir(t, dir)
		run(t, dir, "git", "checkout", "-b", "force/1")
		run(t, dir, "git", "checkout", "-b", "force/2")
		run(t, dir, "git", "checkout", "main")

		// Reset flag before test
		forceFlag = false

		// Parse --force flag
		args := parseGlobalFlags([]string{"--force", "delete", "force"})
		if !forceFlag {
			t.Fatal("--force flag not parsed")
		}
		if len(args) != 2 || args[0] != "delete" || args[1] != "force" {
			t.Fatalf("args not filtered correctly: %v", args)
		}

		// Should not prompt
		err := cmdDelete(args[1:])
		if err != nil {
			t.Fatal(err)
		}

		remaining := branchList(t, dir, "force/*")
		if len(remaining) > 0 {
			t.Fatalf("expected all branches deleted, got: %v", remaining)
		}

		// Reset for other tests
		forceFlag = false
	})

	t.Run("delete with -f", func(t *testing.T) {
		dir := initTestRepo(t)
		inDir(t, dir)
		run(t, dir, "git", "checkout", "-b", "short/1")
		run(t, dir, "git", "checkout", "main")

		forceFlag = false
		args := parseGlobalFlags([]string{"-f", "delete", "short"})
		if !forceFlag {
			t.Fatal("-f flag not parsed")
		}

		err := cmdDelete(args[1:])
		if err != nil {
			t.Fatal(err)
		}

		remaining := branchList(t, dir, "short/*")
		if len(remaining) > 0 {
			t.Fatalf("expected all branches deleted, got: %v", remaining)
		}

		forceFlag = false
	})

	t.Run("delete-upto with --force", func(t *testing.T) {
		dir := initTestRepo(t)
		inDir(t, dir)
		run(t, dir, "git", "checkout", "-b", "upto/1")
		run(t, dir, "git", "checkout", "-b", "upto/2")
		run(t, dir, "git", "checkout", "-b", "upto/3")
		run(t, dir, "git", "checkout", "main")

		forceFlag = false
		args := parseGlobalFlags([]string{"--force", "delete-upto", "upto", "3"})
		if !forceFlag {
			t.Fatal("--force flag not parsed")
		}

		err := cmdDeleteUpto(args[1:])
		if err != nil {
			t.Fatal(err)
		}

		remaining := branchList(t, dir, "upto/*")
		if len(remaining) != 1 || remaining[0] != "upto/3" {
			t.Fatalf("expected only upto/3, got: %v", remaining)
		}

		forceFlag = false
	})

	t.Run("rename with --force", func(t *testing.T) {
		dir := initTestRepo(t)
		inDir(t, dir)
		run(t, dir, "git", "checkout", "-b", "old/1")
		run(t, dir, "git", "checkout", "main")

		forceFlag = false
		args := parseGlobalFlags([]string{"--force", "rename", "old", "new"})
		if !forceFlag {
			t.Fatal("--force flag not parsed")
		}

		err := cmdRename(args[1:])
		if err != nil {
			t.Fatal(err)
		}

		old := branchList(t, dir, "old/*")
		new := branchList(t, dir, "new/*")
		if len(old) > 0 {
			t.Fatalf("expected old/* deleted, got: %v", old)
		}
		if len(new) != 1 || new[0] != "new/1" {
			t.Fatalf("expected new/1, got: %v", new)
		}

		forceFlag = false
	})
}
