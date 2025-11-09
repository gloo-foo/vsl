// Package mount provides utilities for parsing and creating mount specifications.
package mount

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/gloo-foo/vsl/internal/container"
)

// ParseVolume parses a volume specification string (source:target[:ro]) and creates a mount.
// Returns nil if the source path doesn't exist or the specification is invalid.
func ParseVolume(vol container.Volume) *mount.Mount {
	parts := strings.Split(string(vol), ":")
	if len(parts) < 2 {
		return nil
	}

	source := parts[0]
	target := parts[1]
	readonly := len(parts) == 3 && parts[2] == "ro"

	source = expandPath(source)

	// Check if source exists
	if _, err := os.Stat(source); err != nil {
		return nil
	}

	return &mount.Mount{
		Type:     mount.TypeBind,
		Source:   source,
		Target:   target,
		ReadOnly: readonly,
	}
}

// expandPath expands ~ and relative paths to absolute paths.
func expandPath(path string) string {
	// Expand ~/ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}

	// Convert relative to absolute
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			return abs
		}
	}

	return path
}
