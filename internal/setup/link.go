package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkPaths creates symbolic links from source to destination
func LinkPaths(srcBase, dstBase string, paths []string) error {
	for _, p := range paths {
		src := filepath.Join(srcBase, p)
		dst := filepath.Join(dstBase, p)

		// Check if source exists
		if _, err := os.Stat(src); os.IsNotExist(err) {
			fmt.Printf("  skip (not found): %s\n", p)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", p, err)
		}

		// Remove existing destination if it exists
		if _, err := os.Lstat(dst); err == nil {
			if err := os.RemoveAll(dst); err != nil {
				return fmt.Errorf("failed to remove existing %s: %w", p, err)
			}
		}

		// Create symbolic link
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", p, err)
		}
		fmt.Printf("  linked: %s -> %s\n", p, src)
	}
	return nil
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// ReadSymlink returns the target of a symbolic link
func ReadSymlink(path string) (string, error) {
	return os.Readlink(path)
}
