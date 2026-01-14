package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/wt/internal/git"
	"github.com/user/wt/internal/tui"
)

var (
	selectPrintPath bool
)

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively select a worktree",
	Long: `Open an interactive TUI to select an existing worktree.

This is useful for quickly switching between worktrees.
Use --print-path to print the selected path (for shell integration).`,
	RunE: runSelect,
}

func init() {
	selectCmd.Flags().BoolVar(&selectPrintPath, "print-path", false, "Print selected worktree path")
	rootCmd.AddCommand(selectCmd)
}

func runSelect(cmd *cobra.Command, args []string) error {
	repo, err := git.DetectRepository()
	if err != nil {
		return err
	}

	manager := git.NewManager(repo)
	worktrees, err := manager.List()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return nil
	}

	var items []tui.Item
	for _, wt := range worktrees {
		status := ""
		if wt.IsCurrent {
			status = " (current)"
		}
		items = append(items, tui.Item{
			Name:        wt.Branch + status,
			Path:        wt.Path,
			Description: wt.Path,
			IsCurrent:   wt.IsCurrent,
		})
	}

	selected, err := tui.SelectWorktree(items)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil
	}

	if selectPrintPath {
		fmt.Println(selected.Path)
	} else {
		fmt.Printf("Selected: %s\n", selected.Path)
		fmt.Printf("  cd %s\n", selected.Path)
	}

	return nil
}
