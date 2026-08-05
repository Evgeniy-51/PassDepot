package gitremote

import "errors"

var (
	// ErrGitNotFound — git.exe не найден в PATH.
	ErrGitNotFound = errors.New("gitremote: git executable not found in PATH")
)
