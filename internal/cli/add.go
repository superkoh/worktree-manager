package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/superkoh/worktree-manager/internal/config"
	"github.com/superkoh/worktree-manager/internal/git"
	"github.com/superkoh/worktree-manager/internal/setup"
	"github.com/superkoh/worktree-manager/internal/tui"
)

var (
	addNewBranch bool
	addNoSetup   bool
	addPrintPath bool
)

var addCmd = &cobra.Command{
	Use:   "add [branch]",
	Short: "Create a new worktree",
	Long: `Create a new worktree for the specified branch.

If no branch is specified, an interactive selector will be shown.
Use -b to create a new branch.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVarP(&addNewBranch, "new-branch", "b", false, "Create a new branch")
	addCmd.Flags().BoolVar(&addNoSetup, "no-setup", false, "Skip copy/link setup")
	addCmd.Flags().BoolVar(&addPrintPath, "print-path", false, "Print worktree path (for shell integration)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Detect repository
	repo, err := git.DetectRepository()
	if err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.Load(repo.RootPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var branch string

	// Get branch from args or TUI
	if len(args) > 0 {
		branch = args[0]
	} else {
		// Show TUI to select branch
		branches, err := repo.ListBranches()
		if err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}

		// Add remote branches
		remoteBranches, _ := repo.ListRemoteBranches()

		// Create items for TUI
		var items []tui.Item
		for _, b := range branches {
			items = append(items, tui.Item{
				Name:        b,
				Description: "local",
			})
		}
		for _, b := range remoteBranches {
			// Skip if already in local branches
			isLocal := false
			for _, lb := range branches {
				if lb == b {
					isLocal = true
					break
				}
			}
			if !isLocal {
				items = append(items, tui.Item{
					Name:        b,
					Description: "remote",
				})
			}
		}

		selected, err := tui.SelectBranch(items)
		if err != nil {
			return err
		}
		if selected == nil {
			return fmt.Errorf("no branch selected")
		}
		branch = selected.Name
	}

	// Generate worktree path
	basedir, err := cfg.GetWorktreeBasedir(repo.RootPath)
	if err != nil {
		return fmt.Errorf("failed to get basedir: %w", err)
	}

	worktreeName := cfg.GenerateWorktreeName(repo.Name, branch)
	worktreePath := filepath.Join(basedir, worktreeName)

	// Create worktree
	manager := git.NewManager(repo)

	if !addPrintPath {
		fmt.Printf("Creating worktree at: %s\n", worktreePath)
	}

	if err := manager.Add(worktreePath, branch, addNewBranch); err != nil {
		return err
	}

	// Run setup (copy/link)
	if !addNoSetup {
		if err := setup.RunSetup(cfg, repo.RootPath, worktreePath); err != nil {
			// Don't fail, just warn
			fmt.Printf("Warning: setup failed: %v\n", err)
		}
	}

	// Print path for shell integration
	if addPrintPath {
		fmt.Println(worktreePath)
	} else {
		fmt.Printf("\nWorktree created successfully!\n")
		fmt.Printf("  cd %s\n", worktreePath)
	}

	return nil
}
