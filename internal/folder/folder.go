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

// Number extracts the numeric suffix from a branch, or -1 if not numeric.
func Number(branch string) int {
	rhs := branch[strings.Index(branch, "/")+1:]
	n, err := strconv.Atoi(rhs)
	if err != nil || n < 0 {
		return -1
	}
	return n
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

// LastNumber finds the highest numbered branch in a folder.
func LastNumber(folder string) (int, error) {
	branches, err := Enumerate(folder)
	if err != nil {
		return -1, err
	}

	max := -1
	for _, b := range branches {
		n := Number(b)
		if n > max {
			max = n
		}
	}

	if max == -1 {
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

func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func gitRun(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
