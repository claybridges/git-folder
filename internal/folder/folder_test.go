package folder

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// --- Pure function tests ---

func TestIsValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"foo/1", true},
		{"foo/bar", true},
		{"a/b", true},
		{"foo", false},
		{"", false},
		{"/foo", false},
		{"foo/", false},
		{"foo/bar/baz", false},
	}
	for _, tt := range tests {
		if got := IsValid(tt.input); got != tt.want {
			t.Errorf("IsValid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"foo/1", "foo"},
		{"foo/bar", "foo"},
		{"a/b", "a"},
		{"long-name/99", "long-name"},
	}
	for _, tt := range tests {
		if got := Name(tt.input); got != tt.want {
			t.Errorf("Name(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNumber(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"foo/1", 1},
		{"foo/0", 0},
		{"foo/42", 42},
		{"foo/bar", -1},
		{"foo/-1", -1},
		{"foo/1.5", -1},
	}
	for _, tt := range tests {
		if got := Number(tt.input); got != tt.want {
			t.Errorf("Number(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestNumberFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
		ok    bool
	}{
		{"foo/1", 1, true},
		{"foo/0", 0, true},
		{"foo/42", 42, true},
		{"foo/2.5", 2.5, true},
		{"foo/0.1", 0.1, true},
		{"foo/bar", 0, false},
		{"foo/-1", 0, false},
		{"foo/abc", 0, false},
		{"foo/2.50", 0, false},
		{"foo/3.0", 0, false},
		{"foo/10.10", 0, false},
	}
	for _, tt := range tests {
		got, ok := NumberFloat(tt.input)
		if ok != tt.ok || got != tt.want {
			t.Errorf("NumberFloat(%q) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

// --- Git-dependent tests ---

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	return dir
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func inDir(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func TestEnumerate(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/1")
	run(t, dir, "git", "branch", "foo/2")
	run(t, dir, "git", "branch", "foo/3")
	run(t, dir, "git", "branch", "bar/1")

	branches, err := Enumerate("foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 3 {
		t.Fatalf("got %d branches, want 3", len(branches))
	}

	// Should not include bar
	for _, b := range branches {
		if Name(b) != "foo" {
			t.Errorf("unexpected branch %q in foo/ enumeration", b)
		}
	}
}

func TestEnumerateEmpty(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	branches, err := Enumerate("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if branches != nil {
		t.Fatalf("got %v, want nil", branches)
	}
}

func TestLastNumber(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/1")
	run(t, dir, "git", "branch", "foo/5")
	run(t, dir, "git", "branch", "foo/3")

	last, err := LastNumber("foo")
	if err != nil {
		t.Fatal(err)
	}
	if last != 5 {
		t.Errorf("got %v, want 5", last)
	}
}

func TestLastNumberFloat(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/1")
	run(t, dir, "git", "branch", "foo/2.5")

	last, err := LastNumber("foo")
	if err != nil {
		t.Fatal(err)
	}
	if last != 2.5 {
		t.Errorf("got %v, want 2.5", last)
	}
}

func TestLastNumberIntAndFloat(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/1")
	run(t, dir, "git", "branch", "foo/2.5")
	run(t, dir, "git", "branch", "foo/3")

	last, err := LastNumber("foo")
	if err != nil {
		t.Fatal(err)
	}
	if last != 3 {
		t.Errorf("got %v, want 3", last)
	}
}

func TestLastNumberNoNumbered(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "branch", "foo/bar")

	_, err := LastNumber("foo")
	if err == nil {
		t.Fatal("expected error for non-numbered branches")
	}
}

func TestLastNumberNoFolder(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	_, err := LastNumber("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent folder")
	}
}

func TestCurrentFolder(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "-b", "topic/3")

	cur, err := CurrentFolder()
	if err != nil {
		t.Fatal(err)
	}
	if cur != "topic" {
		t.Errorf("got %q, want %q", cur, "topic")
	}
}

func TestCurrentFolderNonFolder(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	cur, err := CurrentFolder()
	if err != nil {
		t.Fatal(err)
	}
	if cur != "" {
		t.Errorf("got %q, want empty", cur)
	}
}

func TestCurrentFolderDetached(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "--detach")

	cur, err := CurrentFolder()
	if err != nil {
		t.Fatal(err)
	}
	if cur != "" {
		t.Errorf("got %q, want empty on detached HEAD", cur)
	}
}

func TestCurrentNumber(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "-b", "topic/3")

	n := CurrentNumber()
	if n != 3 {
		t.Errorf("got %d, want 3", n)
	}
}

func TestCurrentNumberNonFolder(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	n := CurrentNumber()
	if n != -1 {
		t.Errorf("got %d, want -1", n)
	}
}

func TestCurrentNumberDetached(t *testing.T) {
	dir := initTestRepo(t)
	inDir(t, dir)

	run(t, dir, "git", "checkout", "--detach")

	n := CurrentNumber()
	if n != -1 {
		t.Errorf("got %d, want -1", n)
	}
}

// --- Mock tests for git error paths ---

func withMockGit(t *testing.T, fn func(args ...string) (string, error)) {
	t.Helper()
	orig := GitRunner
	GitRunner = fn
	t.Cleanup(func() { GitRunner = orig })
}

func failGit(args ...string) (string, error) {
	return "", fmt.Errorf("mock git failure")
}

func TestEnumerateGitError(t *testing.T) {
	withMockGit(t, failGit)
	_, err := Enumerate("foo")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLastNumberGitError(t *testing.T) {
	withMockGit(t, failGit)
	_, err := LastNumber("foo")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCurrentFolderGitError(t *testing.T) {
	withMockGit(t, failGit)
	cur, err := CurrentFolder()
	// git error on symbolic-ref is treated as "not on a branch"
	if err != nil {
		t.Fatal(err)
	}
	if cur != "" {
		t.Errorf("got %q, want empty on git error", cur)
	}
}
