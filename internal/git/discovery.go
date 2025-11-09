// Package git provides utilities for git repository discovery and management.
package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gloo-foo/vsl/internal/container"
)

// FindRoot finds the root directory of a git repository starting from startDir.
// It walks up the directory tree until it finds a .git directory or reaches the filesystem root.
func FindRoot(startDir string) (container.GitRoot, error) {
	dir := startDir
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return container.GitRoot(dir), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no git repository found")
		}
		dir = parent
	}
}

// FindRealGitDir finds the actual .git directory, resolving worktree references if necessary.
// Git worktrees use a .git file that points to the real git directory.
func FindRealGitDir(gitRoot container.GitRoot) (container.GitDir, error) {
	gitPath := filepath.Join(string(gitRoot), ".git")

	info, err := os.Stat(gitPath)
	if err != nil {
		return "", err
	}

	// If .git is a directory, return it directly
	if info.IsDir() {
		return container.GitDir(gitPath), nil
	}

	// If .git is a file, read it to find the real git directory
	file, err := os.Open(gitPath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "gitdir: ") {
			worktreeGitDir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))

			// For worktrees, we want the main git directory, not the worktree-specific one
			if strings.Contains(worktreeGitDir, "/worktrees/") {
				parts := strings.Split(worktreeGitDir, "/worktrees/")
				if len(parts) > 0 {
					return container.GitDir(parts[0]), nil
				}
			}

			return container.GitDir(worktreeGitDir), nil
		}
	}

	return container.GitDir(gitPath), nil
}
