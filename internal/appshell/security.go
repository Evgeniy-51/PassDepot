package appshell

import (
	"crypto/subtle"
	"errors"
	"strings"

	"passdepot/internal/credstore"
	"passdepot/internal/gitremote"
	"passdepot/internal/profile"
	"passdepot/internal/vaultcore"
)

func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

// ChangeMasterPassword перешифровывает vault и отправляет в репозиторий.
func (a *App) ChangeMasterPassword(oldPassword, newPassword string) error {
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)
	if oldPassword == "" || newPassword == "" {
		return errors.New(L("укажите старый и новый пароль", "Enter both the old and new password"))
	}
	if len(newPassword) < 8 {
		return errors.New(L("новый пароль: минимум 8 символов", "New password must be at least 8 characters"))
	}

	a.mu.Lock()
	if a.profileID == "" || a.vault == nil {
		a.mu.Unlock()
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	if !constantTimeEqual(a.masterPw, []byte(oldPassword)) {
		a.touchActivityLocked()
		a.mu.Unlock()
		return errors.New(L("неверный текущий пароль", "Incorrect current password"))
	}

	p, ok := profile.Get(a.profileID)
	if !ok {
		a.touchActivityLocked()
		a.mu.Unlock()
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	if p.LocalOnly {
		newPw := []byte(newPassword)
		blob, err := vaultcore.EncryptVault(a.vault, newPw)
		if err != nil {
			zeroBytes(newPw)
			a.touchActivityLocked()
			a.mu.Unlock()
			return err
		}
		vp, err := profile.LocalVaultPath(a.profileID)
		if err != nil {
			zeroBytes(newPw)
			a.touchActivityLocked()
			a.mu.Unlock()
			return err
		}
		if err := atomicWriteFile(vp, blob); err != nil {
			zeroBytes(newPw)
			a.touchActivityLocked()
			a.mu.Unlock()
			return err
		}
		zeroBytes(a.masterPw)
		a.masterPw = append([]byte(nil), newPw...)
		zeroBytes(newPw)
		a.dirty = false
		a.entryDirty = false
		a.pendingSync = false
		a.lastErr = ""
		a.touchActivityLocked()
		a.mu.Unlock()
		return nil
	}
	pat, err := credstore.GetPAT(a.profileID)
	if err != nil || pat == "" {
		a.touchActivityLocked()
		a.mu.Unlock()
		return errors.New(L("PAT не найден", "PAT not found"))
	}
	repoDir, err := profile.LocalRepoDir(a.profileID)
	if err != nil {
		a.touchActivityLocked()
		a.mu.Unlock()
		return err
	}
	vp, err := vaultLocalPath(a.profileID)
	if err != nil {
		a.touchActivityLocked()
		a.mu.Unlock()
		return err
	}
	rel := profile.VaultPathInRepo(p)

	newPw := []byte(newPassword)
	blob, err := vaultcore.EncryptVault(a.vault, newPw)
	if err != nil {
		a.touchActivityLocked()
		a.mu.Unlock()
		return err
	}
	if err := atomicWriteFile(vp, blob); err != nil {
		a.touchActivityLocked()
		a.mu.Unlock()
		return err
	}
	if err := gitremote.Add(repoDir, rel); err != nil {
		a.touchActivityLocked()
		a.mu.Unlock()
		return err
	}
	if err := gitremote.Commit(repoDir, commitMessage()); err != nil {
		low := strings.ToLower(err.Error())
		if !strings.Contains(low, "nothing to commit") {
			a.touchActivityLocked()
			a.mu.Unlock()
			return err
		}
	}

	// Локальные изменения (ре-шифрование + commit) уже сделаны. Переключаем мастер-пароль
	// сразу, а push выносим из-под mutex, чтобы не блокировать остальные backend-вызовы.
	a.pendingSync = true
	a.lastErr = ""
	zeroBytes(a.masterPw)
	a.masterPw = append([]byte(nil), newPw...)
	for i := range newPw {
		newPw[i] = 0
	}
	a.dirty = false
	a.entryDirty = false
	a.touchActivityLocked()
	a.mu.Unlock()

	if err := gitremote.Push(pat, repoDir, p.Branch); err != nil {
		a.mu.Lock()
		a.pendingSync = true
		a.lastErr = err.Error()
		a.touchActivityLocked()
		a.mu.Unlock()
		return err
	}

	a.mu.Lock()
	a.pendingSync = false
	a.lastErr = ""
	a.touchActivityLocked()
	a.mu.Unlock()
	return nil
}
