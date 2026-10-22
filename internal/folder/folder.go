package folder

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
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

var numericSuffix = regexp.MustCompile(`^\d+(\.\d*[1-9])?$`)

// NumberFloat extracts a numeric suffix (int or float) from a branch.
// Returns the number and true if the suffix is a plain decimal, or 0 and false otherwise.
// Only `^\d+(\.\d*[1-9])?$` counts: no signs, exponents, hex, "inf", or trailing zeros (e.g. "2.50", "3.0"),
// which keeps each number's representation unique.
func NumberFloat(branch string) (float64, bool) {
	rhs := branch[strings.Index(branch, "/")+1:]
	if !numericSuffix.MatchString(rhs) {
		return 0, false
	}
	n, err := strconv.ParseFloat(rhs, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// Enumerate lists all branches matching a folder prefix.
func Enumerate(folder string) ([]string, error) {
	out, err := git("for-each-ref", "--format=%(refname:short)", "refs/heads/"+folder+"/")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// maxNumbered finds the highest numbered branch in a folder (includes floats), returning its name and number.
func maxNumbered(folder string) (string, float64, error) {
	branches, err := Enumerate(folder)
	if err != nil {
		return "", -1, err
	}

	best := ""
	max := -1.0
	for _, b := range branches {
		n, ok := NumberFloat(b)
		if ok && n > max {
			max = n
			best = b
		}
	}

	if best == "" {
		return "", -1, fmt.Errorf("no numbered branches in folder %s/", folder)
	}
	return best, max, nil
}

// LastNumber finds the highest numbered branch in a folder (includes floats).
func LastNumber(folder string) (float64, error) {
	_, n, err := maxNumbered(folder)
	return n, err
}

// MaxBranch returns the full branch name of the max branch in a folder (e.g. "async/4"), as it exists.
func MaxBranch(folderName string) (string, error) {
	b, _, err := maxNumbered(folderName)
	return b, err
}

// CurrentBranch returns the name of the currently checked-out branch, or "" if detached.
func CurrentBranch() (string, error) {
	out, err := git("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || out == "" {
		return "", nil
	}
	return out, nil
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

// DetectTrunk returns the name of the default branch (e.g. "main" or "master").
// Checks remote HEAD first, then falls back to local branch existence.
func DetectTrunk() (string, error) {
	out, err := git("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil && out != "" {
		// "refs/remotes/origin/main" → "main" (handles slashes like "release/2026")
		const prefix = "refs/remotes/origin/"
		name := strings.TrimPrefix(out, prefix)
		if name != "" && name != out {
			if _, err := git("rev-parse", "--verify", name); err == nil {
				return name, nil
			}
			remoteName := "origin/" + name
			if _, err := git("rev-parse", "--verify", remoteName); err == nil {
				return remoteName, nil
			}
		}
	}
	// Fall back to local branches
	if _, err := git("rev-parse", "--verify", "main"); err == nil {
		return "main", nil
	}
	if _, err := git("rev-parse", "--verify", "master"); err == nil {
		return "master", nil
	}
	return "", fmt.Errorf("cannot determine default branch")
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

const lockKey = "folder.locked"

// configExit reports whether err is a git failure with the given exit status.
func configExit(err error, code int) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == code
}

// LockedFolders returns the names of all locked folders. No key set is not an error.
func LockedFolders() ([]string, error) {
	out, err := git("config", "--get-all", lockKey)
	if err != nil {
		if configExit(err, 1) {
			return nil, nil
		}
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// IsLocked reports whether the folder is locked.
func IsLocked(name string) (bool, error) {
	locked, err := LockedFolders()
	if err != nil {
		return false, err
	}
	return slices.Contains(locked, name), nil
}

// Lock marks the folder as locked. Locking a locked folder is a no-op.
func Lock(name string) error {
	already, err := IsLocked(name)
	if err != nil || already {
		return err
	}
	_, err = git("config", "--add", lockKey, name)
	return err
}

// Unlock clears the folder's lock. Unlocking an unlocked folder is a no-op.
func Unlock(name string) error {
	_, err := git("config", "--fixed-value", "--unset", lockKey, name)
	if err != nil && configExit(err, 5) {
		return nil
	}
	return err
}
