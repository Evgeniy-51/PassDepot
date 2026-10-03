package gitremote

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultUserName  = "PassDepot"
	defaultUserEmail = "passdepot@local"
)

// LsRemote проверяет доступ к репозиторию (git ls-remote).
func LsRemote(pat, repoURL string) error {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return fmt.Errorf("gitremote: empty repo url")
	}
	return runGitErr("", pat, "ls-remote", repoURL)
}

// Clone клонирует repoURL в targetDir (как git clone url targetDir).
func Clone(pat, repoURL, targetDir string) error {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return fmt.Errorf("gitremote: empty repo url")
	}
	if targetDir == "" {
		return fmt.Errorf("gitremote: empty target dir")
	}
	parent := filepath.Dir(targetDir)
	base := filepath.Base(targetDir)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return err
	}
	return runGitErr(parent, pat, "clone", repoURL, base)
}

// Fetch выполняет git fetch origin.
func Fetch(pat, repoDir string) error {
	return runGitErr(repoDir, pat, "fetch", "origin")
}

// Checkout переключает локальную ветку.
func Checkout(repoDir, branch string) error {
	return runGitErr(repoDir, "", "checkout", branch)
}

// ResetHard выполняет git reset --hard ref (например origin/main).
func ResetHard(repoDir, ref string) error {
	return runGitErr(repoDir, "", "reset", "--hard", ref)
}

// LocalChangesError — в клоне есть данные, которых нет в origin; reset --hard уничтожил бы их.
type LocalChangesError struct {
	Ahead       int  // локальные коммиты, не отправленные в origin/<branch>
	Behind      int  // коммиты origin/<branch>, которых нет локально
	Uncommitted bool // файл базы изменён, но не закоммичен
}

func (e *LocalChangesError) Error() string {
	return fmt.Sprintf("gitremote: local changes not pushed (ahead %d, behind %d, uncommitted %t)",
		e.Ahead, e.Behind, e.Uncommitted)
}

// LocalChanges сравнивает HEAD с последним полученным origin/<branch> (без сети).
// vaultRel — путь файла базы в репо; пустой — незакоммиченные изменения не проверяются.
// Возвращает nil, nil, если локальных изменений нет.
func LocalChanges(repoDir, branch, vaultRel string) (*LocalChangesError, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		branch = "main"
	}
	var ahead, behind int
	if revExists(repoDir, "HEAD") {
		remoteRef := "refs/remotes/origin/" + branch
		if revExists(repoDir, remoteRef) {
			out, err := runGitStdout(repoDir, "rev-list", "--left-right", "--count", "HEAD..."+remoteRef)
			if err != nil {
				return nil, err
			}
			f := strings.Fields(string(out))
			if len(f) != 2 {
				return nil, fmt.Errorf("gitremote: unexpected rev-list output %q", string(out))
			}
			if ahead, err = strconv.Atoi(f[0]); err != nil {
				return nil, err
			}
			if behind, err = strconv.Atoi(f[1]); err != nil {
				return nil, err
			}
		} else {
			// Ветки в origin нет (пустой remote, первый push не прошёл) — все локальные коммиты не отправлены.
			out, err := runGitStdout(repoDir, "rev-list", "--count", "HEAD")
			if err != nil {
				return nil, err
			}
			if ahead, err = strconv.Atoi(strings.TrimSpace(string(out))); err != nil {
				return nil, err
			}
		}
	}
	uncommitted := false
	if vaultRel = strings.TrimSpace(vaultRel); vaultRel != "" {
		out, err := runGitStdout(repoDir, "status", "--porcelain", "--untracked-files=no", "--", vaultRel)
		if err != nil {
			return nil, err
		}
		uncommitted = strings.TrimSpace(string(out)) != ""
	}
	if ahead == 0 && !uncommitted {
		return nil, nil
	}
	return &LocalChangesError{Ahead: ahead, Behind: behind, Uncommitted: uncommitted}, nil
}

func revExists(repoDir, rev string) bool {
	_, err := runGitStdout(repoDir, "rev-parse", "--verify", "--quiet", rev+"^{commit}")
	return err == nil
}

// Refresh: fetch + checkout branch + reset --hard origin/branch.
// Если в клоне есть неотправленные коммиты или незакоммиченный файл базы (vaultRel),
// reset не выполняется и возвращается *LocalChangesError.
func Refresh(pat, repoDir, branch, vaultRel string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		branch = "main"
	}
	if err := Fetch(pat, repoDir); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	if err := runGitErr(repoDir, "", "checkout", branch); err != nil {
		if err2 := runGitErr(repoDir, "", "checkout", "-b", branch, "origin/"+branch); err2 != nil {
			return fmt.Errorf("checkout %s: %v; fallback: %w", branch, err, err2)
		}
	}
	lc, err := LocalChanges(repoDir, branch, vaultRel)
	if err != nil {
		return fmt.Errorf("local changes: %w", err)
	}
	if lc != nil {
		return lc
	}
	ref := "origin/" + branch
	if err := ResetHard(repoDir, ref); err != nil {
		return fmt.Errorf("reset --hard %s: %w", ref, err)
	}
	return nil
}

// Add выполняет git add -- paths.
func Add(repoDir string, paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	return runGitErr(repoDir, "", args...)
}

// Commit создаёт коммит с локальным user.name/email (без сети).
func Commit(repoDir, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("gitremote: empty commit message")
	}
	gp, err := GitPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(gp,
		"-c", "user.name="+defaultUserName,
		"-c", "user.email="+defaultUserEmail,
		"commit", "-m", message,
	)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit: %w: %s", err, string(out))
	}
	return nil
}

// Push выполняет git push origin branch.
func Push(pat, repoDir, branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		branch = "main"
	}
	return runGitErr(repoDir, pat, "push", "origin", branch)
}

// EnsureBranch переключает на branch или создаёт её (checkout -B).
func EnsureBranch(repoDir, branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		branch = "main"
	}
	if err := Checkout(repoDir, branch); err == nil {
		return nil
	}
	return runGitErr(repoDir, "", "checkout", "-B", branch)
}
