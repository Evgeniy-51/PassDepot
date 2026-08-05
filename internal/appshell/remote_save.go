package appshell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"passdepot/internal/credstore"
	"passdepot/internal/gitremote"
	"passdepot/internal/profile"
	"passdepot/internal/vaultcore"
)

func normalizeRepoURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimRight(u, "/")
	return u
}

func sameRemote(aURL, aBranch, bURL, bBranch string) bool {
	return strings.EqualFold(normalizeRepoURL(aURL), normalizeRepoURL(bURL)) &&
		strings.EqualFold(strings.TrimSpace(aBranch), strings.TrimSpace(bBranch))
}

// SaveProfileRemote обновляет PAT и/или мигрирует vault на другой репозиторий.
//
// Если URL+branch совпадают с текущими — только проверка доступа и сохранение PAT
// (masterPassword не нужен).
//
// Если URL или branch изменились — требуется masterPassword; текущий vault
// копируется (encrypt+push) в новый remote, профиль переключается.
// Старый репозиторий не очищается (история Git сохранит .pd).
func (a *App) SaveProfileRemote(repoURL, branch, pat, masterPassword string) (RemoteSaveResult, error) {
	var zero RemoteSaveResult

	repoURL = strings.TrimSpace(repoURL)
	branch = strings.TrimSpace(branch)
	pat = strings.TrimSpace(pat)
	masterPassword = strings.TrimSpace(masterPassword)

	if repoURL == "" {
		return zero, errors.New(L("укажите URL репозитория", "Enter the repository URL"))
	}
	if !strings.HasPrefix(strings.ToLower(repoURL), "https://") {
		return zero, errors.New(L("URL репозитория должен начинаться с https://", "Repository URL must start with https://"))
	}
	if branch == "" {
		branch = profile.DefaultBranch
	}
	if pat == "" {
		return zero, errors.New(L("укажите GitHub (GitLab) PAT", "Enter the GitHub (GitLab) PAT"))
	}

	a.mu.Lock()
	if a.profileID == "" || a.vault == nil {
		a.mu.Unlock()
		return zero, errors.New(L("не выполнен вход", "Not signed in"))
	}
	profileID := a.profileID
	p, ok := profile.Get(profileID)
	if !ok {
		a.touchActivityLocked()
		a.mu.Unlock()
		return zero, errors.New(L("профиль не найден", "Profile not found"))
	}
	if p.LocalOnly {
		a.touchActivityLocked()
		a.mu.Unlock()
		return zero, errors.New(L("локальному профилю удалённый репозиторий не нужен", "Local profiles do not use a remote repository"))
	}

	oldURL := p.RepoURL
	oldBranch := p.Branch
	if oldBranch == "" {
		oldBranch = profile.DefaultBranch
	}

	if sameRemote(oldURL, oldBranch, repoURL, branch) {
		a.touchActivityLocked()
		a.mu.Unlock()
		if err := gitremote.LsRemote(pat, repoURL); err != nil {
			return zero, fmt.Errorf("%s: %w", L("доступ к репозиторию", "Repository access"), err)
		}
		if err := credstore.SetPAT(profileID, pat); err != nil {
			return zero, err
		}
		return RemoteSaveResult{
			Migrated:   false,
			OldRepoURL: oldURL,
			NewRepoURL: repoURL,
		}, nil
	}

	if !constantTimeEqual(a.masterPw, []byte(masterPassword)) {
		a.touchActivityLocked()
		a.mu.Unlock()
		return zero, errors.New(L("неверный master password", "Incorrect master password"))
	}

	blob, err := vaultcore.EncryptVault(a.vault, a.masterPw)
	if err != nil {
		a.touchActivityLocked()
		a.mu.Unlock()
		return zero, err
	}
	vaultRel := profile.VaultPathInRepo(p)
	a.touchActivityLocked()
	a.mu.Unlock()

	if err := gitremote.LsRemote(pat, repoURL); err != nil {
		return zero, fmt.Errorf("%s: %w", L("доступ к новому репозиторию", "Access to the new repository"), err)
	}

	root, err := profile.DataRoot()
	if err != nil {
		return zero, err
	}
	tmpDir := filepath.Join(root, "repos", profileID+"-migrate-tmp")
	_ = os.RemoveAll(tmpDir)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := gitremote.Clone(pat, repoURL, tmpDir); err != nil {
		return zero, fmt.Errorf("%s: %w", L("clone нового репозитория", "Cloning the new repository"), err)
	}
	if err := gitremote.EnsureBranch(tmpDir, branch); err != nil {
		return zero, fmt.Errorf("%s: %w", Lf("ветка %s", "branch %s", branch), err)
	}

	vp := filepath.Join(tmpDir, filepath.FromSlash(vaultRel))
	if err := atomicWriteFile(vp, blob); err != nil {
		return zero, err
	}
	if err := gitremote.Add(tmpDir, vaultRel); err != nil {
		return zero, err
	}
	if err := gitremote.Commit(tmpDir, commitMessage()); err != nil {
		low := strings.ToLower(err.Error())
		if !strings.Contains(low, "nothing to commit") {
			return zero, err
		}
	}
	if err := gitremote.Push(pat, tmpDir, branch); err != nil {
		return zero, fmt.Errorf("%s: %w", L("push в новый репозиторий", "Push to the new repository"), err)
	}

	// Успешный push — переключаем профиль и локальный клон.
	p.RepoURL = repoURL
	p.Branch = branch
	p.Normalize()
	if err := p.Validate(); err != nil {
		return zero, err
	}
	if err := profile.Update(p); err != nil {
		return zero, fmt.Errorf("%s: %w", L("обновление профиля (vault уже в новом репо)", "Updating profile (vault already migrated to new repo)"), mapProfileErr(err))
	}
	if err := credstore.SetPAT(profileID, pat); err != nil {
		return zero, fmt.Errorf("%s: %w", L("сохранение PAT (vault уже в новом репо)", "Saving PAT (vault already migrated to new repo)"), err)
	}

	finalDir, err := profile.LocalRepoDir(profileID)
	if err != nil {
		return zero, err
	}
	_ = os.RemoveAll(finalDir)
	if err := os.Rename(tmpDir, finalDir); err != nil {
		// fallback: clone again into final
		_ = os.RemoveAll(finalDir)
		if err2 := gitremote.Clone(pat, repoURL, finalDir); err2 != nil {
			return zero, fmt.Errorf("%s: %v; rename: %w", L("локальный клон после migrate", "local clone after migrate"), err2, err)
		}
		if err2 := gitremote.EnsureBranch(finalDir, branch); err2 != nil {
			return zero, err2
		}
		vp2 := filepath.Join(finalDir, filepath.FromSlash(vaultRel))
		if err2 := atomicWriteFile(vp2, blob); err2 != nil {
			return zero, err2
		}
	}

	a.mu.Lock()
	if a.profileID == profileID {
		a.dirty = false
		a.entryDirty = false
		a.pendingSync = false
		a.lastErr = ""
		a.touchActivityLocked()
	}
	a.mu.Unlock()

	return RemoteSaveResult{
		Migrated:   true,
		OldRepoURL: oldURL,
		NewRepoURL: repoURL,
	}, nil
}
