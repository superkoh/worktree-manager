package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/superkoh/worktree-manager/internal/git"
)

var (
	listJSON bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	Long:    `List all worktrees in the current repository.`,
	RunE:    runList,
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	repo, err := git.DetectRepository()
	if err != nil {
		return err
	}

	manager := git.NewManager(repo)
	worktrees, err := manager.List()
	if err != nil {
		return err
	}

	if listJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(worktrees)
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BRANCH\tPATH\tSTATUS")
	fmt.Fprintln(w, "------\t----\t------")

	for _, wt := range worktrees {
		status := ""
		if wt.IsCurrent {
			status = "* current"
		} else if wt.IsLocked {
			status = "locked"
		} else if wt.IsPrunable {
			status = "prunable"
		} else if wt.IsBare {
			status = "bare"
		}

		branch := wt.Branch
		if branch == "" {
			branch = wt.Head[:7] // Show short commit hash
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", branch, wt.Path, status)
	}

	return w.Flush()
}
