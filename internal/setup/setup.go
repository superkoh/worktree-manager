package setup

import (
	"fmt"

	"github.com/superkoh/worktree-manager/internal/config"
)

// RunSetup performs the copy and link operations for a new worktree
func RunSetup(cfg *config.Config, srcDir, dstDir string, quiet bool) error {
	// Copy files
	if len(cfg.Setup.Copy) > 0 {
		if !quiet {
			fmt.Println("Copying files...")
		}
		if err := CopyFiles(srcDir, dstDir, cfg.Setup.Copy, quiet); err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}
	}

	// Create symlinks
	if len(cfg.Setup.Link) > 0 {
		if !quiet {
			fmt.Println("Creating symlinks...")
		}
		if err := LinkPaths(srcDir, dstDir, cfg.Setup.Link, quiet); err != nil {
			return fmt.Errorf("link failed: %w", err)
		}
	}

	return nil
}
