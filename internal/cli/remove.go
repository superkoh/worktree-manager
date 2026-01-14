package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/wt/internal/git"
	"github.com/user/wt/internal/tui"
)

var (
	removeForce        bool
	removeDeleteBranch bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [worktree...]",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Long: `Remove one or more worktrees.

If no worktree is specified, an interactive selector will be shown.
Use -f to force removal even if there are uncommitted changes.
Use -D to also delete the associated branch.`,
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Force removal")
	removeCmd.Flags().BoolVarP(&removeDeleteBranch, "delete-branch", "D", false, "Also delete the branch")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	repo, err := git.DetectRepository()
	if err != nil {
		return err
	}

	manager := git.NewManager(repo)

	var paths []string

	if len(args) > 0 {
		paths = args
	} else {
		// Show TUI to select worktree
		worktrees, err := manager.List()
		if err != nil {
			return err
		}

		// Filter out main worktree and current worktree
		var items []tui.Item
		for _, wt := range worktrees {
			if wt.IsCurrent {
				continue
			}
			items = append(items, tui.Item{
				Name:        wt.Branch,
				Path:        wt.Path,
				Description: wt.Path,
			})
		}

		if len(items) == 0 {
			fmt.Println("No worktrees to remove.")
			return nil
		}

		selected, err := tui.SelectWorktree(items)
		if err != nil {
			return err
		}
		if selected == nil {
			return fmt.Errorf("no worktree selected")
		}
		paths = []string{selected.Path}
	}

	for _, path := range paths {
		wt, err := manager.FindByPath(path)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			continue
		}

		branch := wt.Branch
		fmt.Printf("Removing worktree: %s (%s)\n", path, branch)

		if err := manager.Remove(path, removeForce); err != nil {
			return err
		}

		if removeDeleteBranch && branch != "" && branch != "(detached)" {
			fmt.Printf("Deleting branch: %s\n", branch)
			if err := manager.DeleteBranch(branch, removeForce); err != nil {
				fmt.Printf("Warning: failed to delete branch %s: %v\n", branch, err)
			}
		}

		fmt.Println("Done!")
	}

	return nil
}
