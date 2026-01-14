package util

import (
	"fmt"
)

// ErrorCode represents different error types
type ErrorCode int

const (
	ErrNotGitRepo ErrorCode = iota + 1
	ErrBranchNotFound
	ErrWorktreeExists
	ErrWorktreeNotFound
	ErrConfigInvalid
	ErrPermissionDenied
	ErrGitCommand
)

// WTError is a custom error type with error codes
type WTError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *WTError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *WTError) Unwrap() error {
	return e.Cause
}

// Error constructors

func NotGitRepoError() *WTError {
	return &WTError{
		Code:    ErrNotGitRepo,
		Message: "not a git repository (or any of the parent directories)",
	}
}

func BranchNotFoundError(branch string) *WTError {
	return &WTError{
		Code:    ErrBranchNotFound,
		Message: fmt.Sprintf("branch '%s' not found", branch),
	}
}

func WorktreeExistsError(path string) *WTError {
	return &WTError{
		Code:    ErrWorktreeExists,
		Message: fmt.Sprintf("worktree already exists at '%s'", path),
	}
}

func WorktreeNotFoundError(path string) *WTError {
	return &WTError{
		Code:    ErrWorktreeNotFound,
		Message: fmt.Sprintf("worktree not found at '%s'", path),
	}
}

func GitCommandError(cmd string, err error) *WTError {
	return &WTError{
		Code:    ErrGitCommand,
		Message: fmt.Sprintf("git command failed: %s", cmd),
		Cause:   err,
	}
}
