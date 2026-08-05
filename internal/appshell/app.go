package appshell

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"passdepot/internal/credstore"
	"passdepot/internal/gitremote"
	"passdepot/internal/profile"
	"passdepot/internal/vaultcore"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App — Wails backend и сессия vault.
type App struct {
	ctx context.Context

	mu sync.Mutex

	profileID   string
	masterPw    []byte
	vault       *vaultcore.Vault
	dirty       bool
	entryDirty  bool // есть несохранённые в файл изменения записей (не только папок)
	pendingSync bool
	lastErr     string
	lastPullAt  time.Time

	autoLockMinutes int
	lastActivity    time.Time
	autoLockCancel  context.CancelFunc

	// pendingLocalVault — .pd после PickLocalVaultImport, до Confirm/Cancel.
	pendingLocalVault []byte
}

// NewApp создаёт приложение.
func NewApp() *App {
	return &App{autoLockMinutes: defaultAutoLockMinutes}
}

// Startup вызывается при старте Wails.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Окно стартует скрытым (StartHidden). Здесь выставляем финальный размер/позицию и показываем,
	// чтобы не было фликера/«второго окна».
	screens, err := runtime.ScreenGetAll(ctx)
	if err == nil && len(screens) > 0 {
		s := screens[0]
		for _, sc := range screens {
			if sc.IsPrimary {
				s = sc
				break
			}
		}
		w := s.Width / 3
		if w < 980 {
			w = 980
		}
		h := (w * 2) / 3
		if h < 680 {
			h = 680
		}
		runtime.WindowSetSize(ctx, w, h)
	}
	runtime.WindowCenter(ctx)
	runtime.WindowShow(ctx)
}

func (a *App) logoutLocked() {
	a.stopAutoLockLocked()
	if a.ctx != nil {
		_ = runtime.ClipboardSetText(a.ctx, "")
	}
	zeroBytes(a.masterPw)
	a.masterPw = nil
	a.vault = nil
	a.profileID = ""
	a.dirty = false
	a.entryDirty = false
	a.pendingSync = false
	a.lastErr = ""
	a.lastPullAt = time.Time{}
	a.clearPendingLocalVaultLocked()
}

func emptyVault() *vaultcore.Vault {
	return &vaultcore.Vault{Version: 1}
}

func commitMessage() string {
	return "Save " + time.Now().Format("2006-01-02 15:04:05 -0700")
}

func toProfileDTO(p profile.Profile) ProfileDTO {
	p.Normalize()
	dto := ProfileDTO{
		ID:          p.ID,
		DisplayName: p.DisplayName,
		RepoURL:     p.RepoURL,
		Branch:      p.Branch,
		LocalOnly:   p.LocalOnly,
	}
	if p.LocalOnly {
		return dto
	}
	pat, err := credstore.GetPAT(p.ID)
	dto.HasPAT = err == nil && strings.TrimSpace(pat) != ""
	return dto
}

// ListProfiles возвращает сохранённые профили.
func (a *App) ListProfiles() ([]ProfileDTO, error) {
	if err := profile.EnsureDataRoot(); err != nil {
		return nil, err
	}
	list, err := profile.List()
	if err != nil {
		return nil, err
	}
	out := make([]ProfileDTO, 0, len(list))
	for _, p := range list {
		out = append(out, toProfileDTO(p))
	}
	return out, nil
}

// CreateProfile создаёт локальный или Git-профиль. PAT нужен только Git-профилю.
func (a *App) CreateProfile(displayName, repoURL, branch, pat string, localOnly bool) (ProfileDTO, error) {
	var zero ProfileDTO
	if err := profile.EnsureDataRoot(); err != nil {
		return zero, err
	}
	if !localOnly {
		if err := gitremote.LsRemote(pat, repoURL); err != nil {
			return zero, fmt.Errorf("%s: %w", L("доступ к репозиторию", "Repository access"), err)
		}
	} else {
		repoURL = ""
		branch = ""
		pat = ""
	}
	p, err := profile.Add(displayName, repoURL, branch, localOnly)
	if err != nil {
		return zero, mapProfileErr(err)
	}
	if localOnly {
		return toProfileDTO(p), nil
	}
	if err := credstore.SetPAT(p.ID, pat); err != nil {
		_ = profile.Remove(p.ID)
		return zero, fmt.Errorf("%s: %w", L("сохранение PAT", "Saving PAT"), err)
	}
	return toProfileDTO(p), nil
}

// DeleteProfile удаляет профиль локально (конфиг, PAT, клон), без remote.
func (a *App) DeleteProfile(profileID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.profileID == profileID {
		a.logoutLocked()
	}
	_ = credstore.DeletePAT(profileID)
	_ = profile.RemoveLocalRepo(profileID)
	_ = profile.RemoveLocalVault(profileID)
	return profile.Remove(profileID)
}

// GetSession возвращает текущее состояние сессии.
func (a *App) GetSession() SessionDTO {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.profileID == "" || a.vault == nil {
		return SessionDTO{}
	}
	p, ok := profile.Get(a.profileID)
	display := ""
	repoURL := ""
	branch := ""
	localOnly := false
	if ok {
		display = p.DisplayName
		repoURL = p.RepoURL
		branch = p.Branch
		localOnly = p.LocalOnly
	}
	lastPull := ""
	if !a.lastPullAt.IsZero() {
		lastPull = a.lastPullAt.Format(time.RFC3339)
	}
	return SessionDTO{
		Unlocked:        true,
		ProfileID:       a.profileID,
		DisplayName:     display,
		RepoURL:         repoURL,
		Branch:          branch,
		Dirty:           a.dirty,
		EntryDirty:      a.entryDirty,
		PendingSync:     a.pendingSync,
		LocalOnly:       localOnly,
		LastError:       a.lastErr,
		AutoLockMinutes: a.autoLockMinutes,
		LastPullAt:      lastPull,
	}
}

// GitVersion возвращает вывод git --version или ошибку.
func (a *App) GitVersion() (string, error) {
	return gitremote.Version()
}

// Login клонирует/обновляет репо и открывает vault. Git-операции выполняются без удержания mutex.
func (a *App) Login(profileID string, masterPassword string) error {
	if strings.TrimSpace(profileID) == "" || masterPassword == "" {
		return errors.New(L("укажите профиль и master password", "Select a profile and enter the master password"))
	}

	a.mu.Lock()
	a.logoutLocked()
	a.mu.Unlock()

	p, ok := profile.Get(profileID)
	if !ok {
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	if p.LocalOnly {
		return a.loginLocal(profileID, masterPassword)
	}
	pat, err := credstore.GetPAT(profileID)
	if err != nil || strings.TrimSpace(pat) == "" {
		return errors.New(L("PAT не найден; укажите токен в настройках", "PAT not found; enter the token in profile settings"))
	}

	repoDir, err := profile.LocalRepoDir(profileID)
	if err != nil {
		return err
	}

	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := gitremote.Clone(pat, p.RepoURL, repoDir); err != nil {
			return fmt.Errorf("%s: %w", L("clone", "Clone"), err)
		}
	} else if err != nil {
		return err
	} else {
		if err := gitremote.Refresh(pat, repoDir, p.Branch); err != nil {
			return fmt.Errorf("%s: %w", L("refresh", "Refresh"), err)
		}
	}

	mp := []byte(masterPassword)
	masterCopy := append([]byte(nil), mp...)
	for i := range mp {
		mp[i] = 0
	}

	a.mu.Lock()

	a.masterPw = masterCopy

	vp, err := vaultRepoPath(profileID)
	if err != nil {
		a.logoutLocked()
		a.mu.Unlock()
		return err
	}

	rel := profile.VaultPathInRepo(p)

	if _, err := os.Stat(vp); os.IsNotExist(err) {
		v := emptyVault()
		blob, err := vaultcore.EncryptVault(v, a.masterPw)
		if err != nil {
			a.logoutLocked()
			a.mu.Unlock()
			return err
		}
		if err := atomicWriteFile(vp, blob); err != nil {
			a.logoutLocked()
			a.mu.Unlock()
			return err
		}
		a.vault = v
		a.profileID = profileID
		a.dirty = false
		a.entryDirty = false
		a.mu.Unlock()

		if err := gitremote.Add(repoDir, rel); err != nil {
			a.mu.Lock()
			a.logoutLocked()
			a.mu.Unlock()
			return err
		}
		if err := gitremote.Commit(repoDir, commitMessage()); err != nil {
			a.mu.Lock()
			a.logoutLocked()
			a.mu.Unlock()
			return err
		}
		if err := gitremote.Push(pat, repoDir, p.Branch); err != nil {
			a.mu.Lock()
			a.pendingSync = true
			a.lastErr = err.Error()
			a.startAutoLockLocked()
			a.mu.Unlock()
			return nil
		}
		a.mu.Lock()
		a.pendingSync = false
		a.lastErr = ""
		a.lastPullAt = time.Now()
		a.startAutoLockLocked()
		a.mu.Unlock()
		return nil
	} else if err != nil {
		a.logoutLocked()
		a.mu.Unlock()
		return err
	}

	data, err := os.ReadFile(vp)
	if err != nil {
		a.logoutLocked()
		a.mu.Unlock()
		return err
	}
	v, err := vaultcore.DecryptVault(data, a.masterPw)
	if err != nil {
		a.logoutLocked()
		a.mu.Unlock()
		return errors.New(L("неверный master password или повреждённый файл", "Incorrect master password or corrupted vault file"))
	}
	a.vault = v
	a.profileID = profileID
	a.dirty = false
	a.entryDirty = false
	a.pendingSync = false
	a.lastErr = ""
	a.lastPullAt = time.Now()
	a.startAutoLockLocked()
	a.mu.Unlock()
	return nil
}

func (a *App) loginLocal(profileID, masterPassword string) error {
	vp, err := profile.LocalVaultPath(profileID)
	if err != nil {
		return err
	}
	masterCopy := []byte(masterPassword)

	var v *vaultcore.Vault
	if _, err := os.Stat(vp); os.IsNotExist(err) {
		v = emptyVault()
		blob, encErr := vaultcore.EncryptVault(v, masterCopy)
		if encErr != nil {
			zeroBytes(masterCopy)
			return encErr
		}
		if writeErr := atomicWriteFile(vp, blob); writeErr != nil {
			zeroBytes(masterCopy)
			return writeErr
		}
	} else if err != nil {
		zeroBytes(masterCopy)
		return err
	} else {
		data, readErr := os.ReadFile(vp)
		if readErr != nil {
			zeroBytes(masterCopy)
			return readErr
		}
		v, err = vaultcore.DecryptVault(data, masterCopy)
		if err != nil {
			zeroBytes(masterCopy)
			return errors.New(L("неверный master password или повреждённый файл", "Incorrect master password or corrupted vault file"))
		}
	}

	a.mu.Lock()
	a.masterPw = masterCopy
	a.vault = v
	a.profileID = profileID
	a.dirty = false
	a.entryDirty = false
	a.pendingSync = false
	a.lastErr = ""
	a.lastPullAt = time.Time{}
	a.startAutoLockLocked()
	a.mu.Unlock()
	return nil
}

// Logout сбрасывает сессию и очищает master password в памяти.
func (a *App) Logout() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logoutLocked()
}

// Save шифрует vault, commit, push. При ошибке push — pendingSync, локальные данные сохранены.
func (a *App) Save() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.profileID == "" || a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	if !a.dirty && !a.pendingSync {
		a.touchActivityLocked()
		return nil
	}

	p, ok := profile.Get(a.profileID)
	if !ok {
		a.touchActivityLocked()
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	if p.LocalOnly {
		if a.dirty {
			blob, err := vaultcore.EncryptVault(a.vault, a.masterPw)
			if err != nil {
				a.touchActivityLocked()
				return err
			}
			vp, err := profile.LocalVaultPath(a.profileID)
			if err != nil {
				a.touchActivityLocked()
				return err
			}
			if err := atomicWriteFile(vp, blob); err != nil {
				a.touchActivityLocked()
				return err
			}
			a.dirty = false
			a.entryDirty = false
		}
		a.pendingSync = false
		a.lastErr = ""
		a.touchActivityLocked()
		return nil
	}
	pat, err := credstore.GetPAT(a.profileID)
	if err != nil || pat == "" {
		a.touchActivityLocked()
		return errors.New(L("PAT не найден", "PAT not found"))
	}
	repoDir, err := profile.LocalRepoDir(a.profileID)
	if err != nil {
		a.touchActivityLocked()
		return err
	}
	vp, err := vaultRepoPath(a.profileID)
	if err != nil {
		a.touchActivityLocked()
		return err
	}
	rel := profile.VaultPathInRepo(p)

	if a.dirty {
		blob, err := vaultcore.EncryptVault(a.vault, a.masterPw)
		if err != nil {
			a.touchActivityLocked()
			return err
		}
		if err := atomicWriteFile(vp, blob); err != nil {
			a.touchActivityLocked()
			return err
		}
		if err := gitremote.Add(repoDir, rel); err != nil {
			a.touchActivityLocked()
			return err
		}
		if err := gitremote.Commit(repoDir, commitMessage()); err != nil {
			low := strings.ToLower(err.Error())
			if !strings.Contains(low, "nothing to commit") {
				a.touchActivityLocked()
				return err
			}
		}
		a.dirty = false
		a.entryDirty = false
	}

	if err := gitremote.Push(pat, repoDir, p.Branch); err != nil {
		a.pendingSync = true
		a.lastErr = err.Error()
		a.touchActivityLocked()
		return err
	}
	a.pendingSync = false
	a.lastErr = ""
	a.touchActivityLocked()
	return nil
}

// Refresh подтягивает remote и перечитывает vault (только если нет несохранённых изменений).
func (a *App) Refresh() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.profileID == "" || a.vault == nil {
		a.lastErr = L("Сессия неактивна; войдите снова.", "Session is inactive; please sign in again.")
		return errors.New(a.lastErr)
	}
	if a.dirty {
		a.lastErr = L("Сначала сохраните изменения, затем обновите с репозитория.", "Save your changes first, then refresh from the repository.")
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}

	p, ok := profile.Get(a.profileID)
	if !ok {
		a.lastErr = L("Профиль не найден.", "Profile not found.")
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}
	if p.LocalOnly {
		a.lastErr = ""
		a.touchActivityLocked()
		return nil
	}
	pat, err := credstore.GetPAT(a.profileID)
	if err != nil || pat == "" {
		a.lastErr = L("PAT не найден; укажите токен в настройках профиля.", "PAT not found; enter the token in profile settings.")
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}
	repoDir, err := profile.LocalRepoDir(a.profileID)
	if err != nil {
		a.lastErr = formatPullError(err)
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}
	if err := gitremote.Refresh(pat, repoDir, p.Branch); err != nil {
		a.lastErr = formatPullError(err)
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}
	vp, err := vaultRepoPath(a.profileID)
	if err != nil {
		a.lastErr = formatPullError(err)
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}
	data, err := os.ReadFile(vp)
	if err != nil {
		a.lastErr = formatPullError(err)
		a.touchActivityLocked()
		return errors.New(a.lastErr)
	}
	v, err := vaultcore.DecryptVault(data, a.masterPw)
	if err != nil {
		msg := L("Не удалось расшифровать базу после обновления (чужой файл или повреждение).", "Failed to decrypt the vault after refresh (unexpected file or corruption).")
		a.lastErr = msg
		a.touchActivityLocked()
		return errors.New(msg)
	}
	a.vault = v
	a.pendingSync = false
	a.lastErr = ""
	a.lastPullAt = time.Now()
	a.dirty = false
	a.entryDirty = false
	a.touchActivityLocked()
	return nil
}

// RetryPush повторяет только push (если был сбой сети).
func (a *App) RetryPush() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.profileID == "" || a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	if !a.pendingSync {
		a.touchActivityLocked()
		return nil
	}
	p, ok := profile.Get(a.profileID)
	if !ok {
		a.touchActivityLocked()
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	if p.LocalOnly {
		a.pendingSync = false
		a.lastErr = ""
		a.touchActivityLocked()
		return nil
	}
	pat, err := credstore.GetPAT(a.profileID)
	if err != nil || pat == "" {
		a.touchActivityLocked()
		return errors.New(L("PAT не найден", "PAT not found"))
	}
	repoDir, err := profile.LocalRepoDir(a.profileID)
	if err != nil {
		a.touchActivityLocked()
		return err
	}
	if err := gitremote.Push(pat, repoDir, p.Branch); err != nil {
		a.lastErr = err.Error()
		a.touchActivityLocked()
		return err
	}
	a.pendingSync = false
	a.lastErr = ""
	a.touchActivityLocked()
	return nil
}
