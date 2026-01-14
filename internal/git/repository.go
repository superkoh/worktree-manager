package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/user/wt/internal/util"
)

// Repository represents a git repository
type Repository struct {
	RootPath string
	Name     string
}

// DetectRepository finds the git repository root from the current directory
func DetectRepository() (*Repository, error) {
	return DetectRepositoryFrom(".")
}

// DetectRepositoryFrom finds the git repository root from a given path
func DetectRepositoryFrom(path string) (*Repository, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return nil, util.NotGitRepoError()
	}

	rootPath := strings.TrimSpace(string(output))
	return &Repository{
		RootPath: rootPath,
		Name:     filepath.Base(rootPath),
	}, nil
}

// ListBranches returns all local branches
func (r *Repository) ListBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, util.GitCommandError("branch --format", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// ListRemoteBranches returns all remote branches
func (r *Repository) ListRemoteBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, util.GitCommandError("branch -r", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			// Remove origin/ prefix
			if strings.HasPrefix(line, "origin/") {
				line = strings.TrimPrefix(line, "origin/")
			}
			if line != "HEAD" {
				branches = append(branches, line)
			}
		}
	}
	return branches, nil
}

// GetCurrentBranch returns the current branch name
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return "", util.GitCommandError("rev-parse --abbrev-ref HEAD", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// BranchExists checks if a branch exists
func (r *Repository) BranchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = r.RootPath
	cmd.Stderr = nil
	cmd.Stdout = nil
	return cmd.Run() == nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (r *Repository) HasUncommittedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// IsGitRepository checks if a directory is inside a git repository
func IsGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	cmd.Stderr = nil
	cmd.Stdout = nil
	return cmd.Run() == nil
}

// GetGitDir returns the .git directory path
func (r *Repository) GetGitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	gitDir := strings.TrimSpace(string(output))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(r.RootPath, gitDir)
	}
	return gitDir, nil
}

// Fetch fetches from remote
func (r *Repository) Fetch() error {
	cmd := exec.Command("git", "fetch", "--all", "--prune")
	cmd.Dir = r.RootPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
