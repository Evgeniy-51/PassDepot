package gitremote

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var (
	gitPathOnce sync.Once
	gitPath     string
	gitPathErr  error
)

// GitPath возвращает путь к git.exe (кэш). Ищет git.exe, затем git.
func GitPath() (string, error) {
	gitPathOnce.Do(func() {
		p, err := exec.LookPath("git.exe")
		if err != nil {
			p, err = exec.LookPath("git")
		}
		if err != nil {
			gitPathErr = ErrGitNotFound
			return
		}
		gitPath = p
	})
	if gitPathErr != nil {
		return "", gitPathErr
	}
	return gitPath, nil
}

// Version выполняет git --version.
func Version() (string, error) {
	gp, err := GitPath()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(gp, "--version")
	applySysProcAttr(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}

// runGit запускает git в каталоге dir (пустой = текущий процесс не меняется).
// pat: если не пусто — добавляется -c http.extraHeader для GitHub HTTPS.
// args — аргументы после git (например "fetch", "origin").
func runGit(dir, pat string, args ...string) ([]byte, error) {
	gp, err := GitPath()
	if err != nil {
		return nil, err
	}
	var full []string
	if pat != "" {
		full = append(full, "-c", "http.extraHeader="+authHeaderGit(pat))
	}
	full = append(full, args...)
	cmd := exec.Command(gp, full...)
	applySysProcAttr(cmd)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		return out, fmt.Errorf("git: %w\n%s", err, string(out))
	}
	return out, nil
}

func runGitErr(dir, pat string, args ...string) error {
	_, err := runGit(dir, pat, args...)
	return err
}
