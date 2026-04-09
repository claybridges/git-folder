// setup-example creates a git repo with the async/ branch folder from the README example.
// Usage: setup-example [dir]
// If dir is not provided, a new repo is created in a temp directory.
// If dir is provided, a new repo is created inside it.
package main

import (
	"fmt"
	"os"

	"github.com/claybridges/git-folder/internal/example"
)

func main() {
	parent := os.TempDir()
	if len(os.Args) > 1 {
		parent = os.Args[1]
	}

	dir, err := os.MkdirTemp(parent, "git-folder-example-*")
	if err != nil {
		fatalf("create dir: %v", err)
	}

	if err := example.InitRepo(dir); err != nil {
		fatalf("init repo: %v", err)
	}
	fmt.Printf("Created repo at %s\n", dir)

	if err := example.Setup(dir); err != nil {
		fatalf("setup: %v", err)
	}
	fmt.Printf("✓ Created async/ folder — run: git -C %s folder list async\n", dir)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "setup-example: "+format+"\n", args...)
	os.Exit(1)
}
