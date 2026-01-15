package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// LinkPaths creates symbolic links from source to destination
// On Windows, if symlink fails (requires admin/dev mode), it falls back to copy
func LinkPaths(srcBase, dstBase string, paths []string) error {
	for _, p := range paths {
		src := filepath.Join(srcBase, p)
		dst := filepath.Join(dstBase, p)

		// Check if source exists
		info, err := os.Stat(src)
		if os.IsNotExist(err) {
			fmt.Printf("  skip (not found): %s\n", p)
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", p, err)
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

		// Try to create symbolic link
		if err := os.Symlink(src, dst); err != nil {
			// On Windows, symlink may fail without admin privileges
			// Fall back to copying
			if runtime.GOOS == "windows" {
				fmt.Printf("  symlink failed, falling back to copy: %s\n", p)
				if info.IsDir() {
					if err := copyDir(src, dst); err != nil {
						return fmt.Errorf("failed to copy directory %s: %w", p, err)
					}
				} else {
					if err := copyFile(src, dst); err != nil {
						return fmt.Errorf("failed to copy file %s: %w", p, err)
					}
				}
				fmt.Printf("  copied (symlink fallback): %s\n", p)
				continue
			}
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

// SupportsSymlinks checks if the current OS/filesystem supports symlinks
func SupportsSymlinks() bool {
	if runtime.GOOS != "windows" {
		return true
	}
	// On Windows, try to create a test symlink
	tmpDir := os.TempDir()
	testSrc := filepath.Join(tmpDir, "wt_symlink_test_src")
	testDst := filepath.Join(tmpDir, "wt_symlink_test_dst")

	// Create test file
	if err := os.WriteFile(testSrc, []byte("test"), 0644); err != nil {
		return false
	}
	defer os.Remove(testSrc)

	// Try symlink
	err := os.Symlink(testSrc, testDst)
	if err == nil {
		os.Remove(testDst)
		return true
	}
	return false
}
