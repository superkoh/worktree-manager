package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/superkoh/worktree-manager/internal/git"
)

var (
	pruneDryRun bool
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stale worktree references",
	Long: `Remove worktree information for worktrees that no longer exist on disk.

Use --dry-run to preview what would be removed.`,
	RunE: runPrune,
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "Preview what would be removed")
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, args []string) error {
	repo, err := git.DetectRepository()
	if err != nil {
		return err
	}

	manager := git.NewManager(repo)

	if pruneDryRun {
		fmt.Println("Dry run - the following would be pruned:")
	}

	pruned, err := manager.Prune(pruneDryRun)
	if err != nil {
		return err
	}

	if len(pruned) == 0 {
		fmt.Println("Nothing to prune.")
	} else {
		for _, line := range pruned {
			fmt.Println(line)
		}
		if !pruneDryRun {
			fmt.Printf("\nPruned %d stale worktree reference(s).\n", len(pruned))
		}
	}

	return nil
}
