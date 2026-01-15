package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/superkoh/worktree-manager/internal/util"
)

// Worktree represents a git worktree
type Worktree struct {
	Path       string
	Branch     string
	Head       string
	IsBare     bool
	IsLocked   bool
	IsPrunable bool
	IsCurrent  bool
}

// Manager handles worktree operations
type Manager struct {
	repo *Repository
}

// NewManager creates a new worktree manager
func NewManager(repo *Repository) *Manager {
	return &Manager{repo: repo}
}

// List returns all worktrees
func (m *Manager) List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = m.repo.RootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, util.GitCommandError("worktree list", err)
	}

	worktrees, err := parseWorktreeList(string(output))
	if err != nil {
		return nil, err
	}

	// Mark current worktree
	cwd, _ := os.Getwd()
	for i := range worktrees {
		if strings.HasPrefix(cwd, worktrees[i].Path) {
			worktrees[i].IsCurrent = true
		}
	}

	return worktrees, nil
}

// Add creates a new worktree
func (m *Manager) Add(path, branch string, createBranch bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check if path already exists
	if util.FileExists(absPath) {
		return util.WorktreeExistsError(absPath)
	}

	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, absPath)
	} else {
		// Check if branch exists
		if !m.repo.BranchExists(branch) {
			return util.BranchNotFoundError(branch)
		}
		args = append(args, absPath, branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repo.RootPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return util.GitCommandError("worktree add", err)
	}

	return nil
}

// Remove removes a worktree
func (m *Manager) Remove(path string, force bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, absPath)

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repo.RootPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return util.GitCommandError("worktree remove", err)
	}

	return nil
}

// Prune removes worktree information for worktrees that are no longer present
func (m *Manager) Prune(dryRun bool) ([]string, error) {
	args := []string{"worktree", "prune"}
	if dryRun {
		args = append(args, "--dry-run", "-v")
	} else {
		args = append(args, "-v")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repo.RootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, util.GitCommandError("worktree prune", err)
	}

	// Parse output for pruned worktrees
	var pruned []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			pruned = append(pruned, line)
		}
	}

	return pruned, nil
}

// GetMainWorktree returns the main (first) worktree
func (m *Manager) GetMainWorktree() (*Worktree, error) {
	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}
	if len(worktrees) == 0 {
		return nil, fmt.Errorf("no worktrees found")
	}
	return &worktrees[0], nil
}

// FindByPath finds a worktree by its path
func (m *Manager) FindByPath(path string) (*Worktree, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Path == absPath {
			return &wt, nil
		}
	}

	return nil, util.WorktreeNotFoundError(absPath)
}

// FindByBranch finds a worktree by its branch name
func (m *Manager) FindByBranch(branch string) (*Worktree, error) {
	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return &wt, nil
		}
	}

	return nil, nil
}

// DeleteBranch deletes a branch
func (m *Manager) DeleteBranch(branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}

	cmd := exec.Command("git", "branch", flag, branch)
	cmd.Dir = m.repo.RootPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// parseWorktreeList parses the porcelain output of git worktree list
func parseWorktreeList(output string) ([]Worktree, error) {
	var worktrees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			current.Branch = branch
		case line == "bare":
			current.IsBare = true
		case line == "locked":
			current.IsLocked = true
		case strings.HasPrefix(line, "prunable"):
			current.IsPrunable = true
		case line == "detached":
			current.Branch = "(detached)"
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, scanner.Err()
}
