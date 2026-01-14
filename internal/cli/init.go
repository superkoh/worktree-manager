package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/wt/internal/config"
	"github.com/user/wt/internal/git"
	"github.com/user/wt/internal/util"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .wt.json configuration",
	Long: `Create a new .wt.json configuration file in the repository root.

The configuration file allows you to customize:
- Worktree base directory and naming convention
- Files to copy to new worktrees
- Paths to symlink to new worktrees`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	repo, err := git.DetectRepository()
	if err != nil {
		return err
	}

	configPath := filepath.Join(repo.RootPath, config.ConfigFileName)

	if util.FileExists(configPath) {
		return fmt.Errorf("%s already exists", config.ConfigFileName)
	}

	cfg := config.DefaultConfig()
	// Add some sensible defaults for common use cases
	cfg.Setup.Copy = []string{".env", ".env.local"}
	cfg.Setup.Link = []string{"node_modules", "vendor"}

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Printf("Created %s\n", configPath)
	fmt.Println("\nEdit this file to customize:")
	fmt.Println("  - worktree.basedir: where to create worktrees")
	fmt.Println("  - worktree.naming: naming template ({repo}-{branch})")
	fmt.Println("  - setup.copy: files to copy to new worktrees")
	fmt.Println("  - setup.link: paths to symlink to new worktrees")

	return nil
}
