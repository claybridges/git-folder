// Package example creates the async/ branch folder from the README example.
package example

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Setup creates the async/ folder structure in dir, which must already be a
// git repo with at least one commit on main.
func Setup(dir string) error {
	// async/1 branches from main at commit 3 (after a, b, c)
	sha3, err := gout(dir, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if err := g(dir, "checkout", "-b", "async/1", sha3); err != nil {
		return err
	}
	for i := 1; i <= 7; i++ {
		if err := commit(dir, fmt.Sprintf("async1-work-%d", i)); err != nil {
			return err
		}
	}

	// main keeps moving (d–g)
	if err := g(dir, "checkout", "main"); err != nil {
		return err
	}
	for _, l := range []string{"d", "e", "f", "g"} {
		if err := commitFile(dir, l+".go", fmt.Sprintf("package main\n\n// %s\n", l), "add "+l+".go"); err != nil {
			return err
		}
	}

	// async/2: squash of async/1 rebased onto main at g
	if err := g(dir, "checkout", "-b", "async/2", "main"); err != nil {
		return err
	}
	if err := g(dir, "merge", "--squash", "async/1"); err != nil {
		return err
	}
	if err := g(dir, "commit", "-m", "async/1 squashed"); err != nil {
		return err
	}

	// async/2.5: experiment off async/2
	if err := g(dir, "checkout", "-b", "async/2.5"); err != nil {
		return err
	}
	for _, msg := range []string{"async25-experiment-1", "async25-experiment-2"} {
		if err := commit(dir, msg); err != nil {
			return err
		}
	}

	// main keeps moving (h–k)
	if err := g(dir, "checkout", "main"); err != nil {
		return err
	}
	for _, l := range []string{"h", "i", "j", "k"} {
		if l == "j" {
			if err := commitFileAs(dir, "j.go", "package main\n\n// j\n", "add j.go", "Buckaroo Bonzai", "bbonzai@example.com"); err != nil {
				return err
			}
			continue
		}
		if err := commitFile(dir, l+".go", fmt.Sprintf("package main\n\n// %s\n", l), "add "+l+".go"); err != nil {
			return err
		}
	}

	// async/3: squash of async/2 rebased onto main at k
	if err := g(dir, "checkout", "-b", "async/3", "main"); err != nil {
		return err
	}
	if err := g(dir, "merge", "--squash", "async/2"); err != nil {
		return err
	}
	if err := g(dir, "commit", "-m", "async/2 squashed"); err != nil {
		return err
	}
	if err := commit(dir, "async3-followup"); err != nil {
		return err
	}

	// main keeps moving (l–o)
	if err := g(dir, "checkout", "main"); err != nil {
		return err
	}
	for _, l := range []string{"l", "m", "n", "o"} {
		if err := commitFile(dir, l+".go", fmt.Sprintf("package main\n\n// %s\n", l), "add "+l+".go"); err != nil {
			return err
		}
	}

	// async/4: squash of async/3 rebased onto main at o
	if err := g(dir, "checkout", "-b", "async/4", "main"); err != nil {
		return err
	}
	if err := g(dir, "merge", "--squash", "async/3"); err != nil {
		return err
	}
	if err := g(dir, "commit", "-m", "async/3 squashed"); err != nil {
		return err
	}
	for _, msg := range []string{"async4-work-1", "async4-work-2"} {
		if err := commit(dir, msg); err != nil {
			return err
		}
	}

	// main keeps moving (p–s)
	if err := g(dir, "checkout", "main"); err != nil {
		return err
	}
	for _, l := range []string{"p", "q", "r", "s"} {
		if err := commitFile(dir, l+".go", fmt.Sprintf("package main\n\n// %s\n", l), "add "+l+".go"); err != nil {
			return err
		}
	}

	// async/5: squash of async/4 rebased onto main at s
	if err := g(dir, "checkout", "-b", "async/5", "main"); err != nil {
		return err
	}
	if err := g(dir, "merge", "--squash", "async/4"); err != nil {
		return err
	}
	if err := g(dir, "commit", "-m", "async/4 squashed"); err != nil {
		return err
	}
	if err := commit(dir, "async5-work-1"); err != nil {
		return err
	}

	// async/bigbooty: experiment off async/3
	if err := g(dir, "checkout", "-b", "async/bigbooty", "async/3"); err != nil {
		return err
	}
	if err := commit(dir, "bigbooty-experiment"); err != nil {
		return err
	}

	// async/temp: scratch off main
	if err := g(dir, "checkout", "-b", "async/temp", "main"); err != nil {
		return err
	}
	if err := commit(dir, "temp-scratch"); err != nil {
		return err
	}

	return g(dir, "checkout", "main")
}

// InitRepo initializes a new git repo in dir with 3 commits on main (a.go, b.go, c.go).
func InitRepo(dir string) error {
	if err := g(dir, "init", "-b", "main"); err != nil {
		return err
	}
	if err := g(dir, "config", "user.email", "jbigbooty@example.com"); err != nil {
		return err
	}
	if err := g(dir, "config", "user.name", "John Bigbooty"); err != nil {
		return err
	}
	for _, l := range []string{"a", "b", "c"} {
		if err := commitFile(dir, l+".go", fmt.Sprintf("package main\n\n// %s\n", l), "add "+l+".go"); err != nil {
			return err
		}
	}
	return nil
}

func commit(dir, msg string) error {
	return commitFile(dir, msg+".txt", msg, msg)
}

func commitFileAs(dir, filename, content, msg, name, email string) error {
	f := filepath.Join(dir, filename)
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	if err := g(dir, "add", filename); err != nil {
		return err
	}
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+name,
		"GIT_AUTHOR_EMAIL="+email,
		"GIT_COMMITTER_NAME="+name,
		"GIT_COMMITTER_EMAIL="+email,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w\n%s", err, out)
	}
	return nil
}

func commitFile(dir, filename, content, msg string) error {
	f := filepath.Join(dir, filename)
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	if err := g(dir, "add", filename); err != nil {
		return err
	}
	return g(dir, "commit", "-m", msg)
}

func g(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}

func gout(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %v: %w", args, err)
	}
	return strings.TrimSpace(string(out)), nil
}
