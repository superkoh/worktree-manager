package setup

import (
	"fmt"

	"github.com/superkoh/worktree-manager/internal/config"
)

// RunSetup performs the copy and link operations for a new worktree
func RunSetup(cfg *config.Config, srcDir, dstDir string) error {
	// Copy files
	if len(cfg.Setup.Copy) > 0 {
		fmt.Println("Copying files...")
		if err := CopyFiles(srcDir, dstDir, cfg.Setup.Copy); err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}
	}

	// Create symlinks
	if len(cfg.Setup.Link) > 0 {
		fmt.Println("Creating symlinks...")
		if err := LinkPaths(srcDir, dstDir, cfg.Setup.Link); err != nil {
			return fmt.Errorf("link failed: %w", err)
		}
	}

	return nil
}
