package folder

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var validPattern = regexp.MustCompile(`^[^/]+/[^/]+$`)

// IsValid checks if a branch name is in folder/suffix format.
func IsValid(branch string) bool {
	return validPattern.MatchString(branch)
}

// Name extracts the folder prefix from a branch (e.g. "foo" from "foo/3").
func Name(branch string) string {
	return branch[:strings.Index(branch, "/")]
}

// Number extracts the integer suffix from a branch, or -1 if not an integer.
func Number(branch string) int {
	rhs := branch[strings.Index(branch, "/")+1:]
	n, err := strconv.Atoi(rhs)
	if err != nil || n < 0 {
		return -1
	}
	return n
}

// NumberFloat extracts a numeric suffix (int or float) from a branch.
// Returns the number and true if parseable, or 0 and false otherwise.
// Rejects trailing zeros (e.g. "2.50", "3.0") to ensure unique representations.
func NumberFloat(branch string) (float64, bool) {
	rhs := branch[strings.Index(branch, "/")+1:]
	if strings.Contains(rhs, ".") && strings.HasSuffix(rhs, "0") {
		return 0, false
	}
	n, err := strconv.ParseFloat(rhs, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// Enumerate lists all branches matching a folder prefix.
func Enumerate(folder string) ([]string, error) {
	out, err := git("branch", "--format=%(refname:short)", "--list", folder+"/*")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// LastNumber finds the highest numbered branch in a folder (includes floats).
func LastNumber(folder string) (float64, error) {
	branches, err := Enumerate(folder)
	if err != nil {
		return -1, err
	}

	max := -1.0
	found := false
	for _, b := range branches {
		n, ok := NumberFloat(b)
		if ok && n > max {
			max = n
			found = true
		}
	}

	if !found {
		return -1, fmt.Errorf("no numbered branches in folder %s/", folder)
	}
	return max, nil
}

// CurrentFolder returns the folder name of the current branch, or "" if not on a folder branch.
func CurrentFolder() (string, error) {
	out, err := git("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || out == "" {
		return "", nil
	}
	if IsValid(out) {
		return Name(out), nil
	}
	return "", nil
}

// CurrentNumber returns the numeric suffix of the current branch, or -1.
func CurrentNumber() int {
	out, _ := git("symbolic-ref", "--quiet", "--short", "HEAD")
	if out == "" || !IsValid(out) {
		return -1
	}
	return Number(out)
}

// GitRunner executes a git command and returns trimmed output.
// Override in tests to inject errors.
var GitRunner = func(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func git(args ...string) (string, error) {
	return GitRunner(args...)
}

