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

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CopyToClipboard копирует текст в системный буфер обмена (Windows).
func (a *App) CopyToClipboard(text string) error {
	if a.ctx == nil {
		return errors.New(L("нет контекста приложения", "Application context is not available"))
	}
	a.mu.Lock()
	a.touchActivityLocked()
	a.mu.Unlock()
	return runtime.ClipboardSetText(a.ctx, text)
}

// UpdateProfilePAT проверяет доступ к репо и сохраняет новый PAT в Credential Manager.
func (a *App) UpdateProfilePAT(profileID, pat string) error {
	p, ok := profile.Get(profileID)
	if !ok {
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	if p.LocalOnly {
		return errors.New(L("локальному профилю PAT не требуется", "Local profiles do not require a PAT"))
	}
	pat = strings.TrimSpace(pat)
	if pat == "" {
		return errors.New(L("пустой PAT", "PAT is empty"))
	}
	if err := gitremote.LsRemote(pat, p.RepoURL); err != nil {
		return fmt.Errorf("%s: %w", L("доступ к репозиторию", "Repository access"), err)
	}
	return credstore.SetPAT(profileID, pat)
}

// RenameProfile меняет только отображаемое имя. VaultFileName / путь .pd не трогает.
func (a *App) RenameProfile(profileID, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return errors.New(L("укажите имя профиля", "Enter the profile name"))
	}
	p, ok := profile.Get(profileID)
	if !ok {
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	p.DisplayName = displayName
	if err := p.Validate(); err != nil {
		return err
	}
	return mapProfileErr(profile.Update(p))
}

// ExportProfileJSON возвращает JSON профиля без секретов (для копирования вручную).
func (a *App) ExportProfileJSON(profileID string) (string, error) {
	p, ok := profile.Get(profileID)
	if !ok {
		return "", errors.New(L("профиль не найден", "Profile not found"))
	}
	b, err := profile.MarshalExportPublic(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ImportProfileJSON добавляет профиль из JSON (без PAT; PAT ввести отдельно).
func (a *App) ImportProfileJSON(jsonStr string) (ProfileDTO, error) {
	var zero ProfileDTO
	p, err := profile.UnmarshalImportPublic([]byte(jsonStr))
	if err != nil {
		return zero, err
	}
	np, err := profile.AddImported(p)
	if err != nil {
		return zero, mapProfileErr(err)
	}
	return toProfileDTO(np), nil
}

// ExportProfileToFile открывает диалог «Сохранить как» и пишет JSON профиля.
func (a *App) ExportProfileToFile(profileID string) error {
	if a.ctx == nil {
		return errors.New(L("нет контекста приложения", "Application context is not available"))
	}
	p, ok := profile.Get(profileID)
	if !ok {
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	data, err := profile.MarshalExportPublic(p)
	if err != nil {
		return err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           L("Экспорт профиля PassDepot", "Export PassDepot Profile"),
		DefaultFilename: profile.MakeExportProfileFileName(p.DisplayName),
		Filters: []runtime.FileFilter{
			{DisplayName: L("JSON (*.json)", "JSON (*.json)"), Pattern: "*.json"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	return os.WriteFile(path, data, 0o600)
}

// ImportProfileFromFile открывает диалог выбора файла и импортирует профиль.
// Если пользователь отменил диалог, возвращается пустой ProfileDTO без ошибки.
func (a *App) ImportProfileFromFile() (ProfileDTO, error) {
	var zero ProfileDTO
	if a.ctx == nil {
		return zero, errors.New(L("нет контекста приложения", "Application context is not available"))
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: L("Импорт профиля PassDepot", "Import PassDepot Profile"),
		Filters: []runtime.FileFilter{
			{DisplayName: L("JSON (*.json)", "JSON (*.json)"), Pattern: "*.json"},
		},
	})
	if err != nil {
		return zero, err
	}
	if path == "" {
		return zero, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	prof, err := profile.UnmarshalImportPublic(data)
	if err != nil {
		return zero, err
	}
	np, err := profile.AddImported(prof)
	if err != nil {
		return zero, mapProfileErr(err)
	}
	return toProfileDTO(np), nil
}

// ExportLocalVaultToFile сохраняет копию зашифрованного .pd (локальный или Git-профиль).
func (a *App) ExportLocalVaultToFile(profileID string) error {
	if a.ctx == nil {
		return errors.New(L("нет контекста приложения", "Application context is not available"))
	}
	p, ok := profile.Get(profileID)
	if !ok {
		return errors.New(L("профиль не найден", "Profile not found"))
	}
	if err := a.flushVaultFile(profileID); err != nil {
		return err
	}
	source, err := vaultLocalPath(profileID)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           L("Экспорт хранилища PassDepot", "Export PassDepot Vault"),
		DefaultFilename: profile.MakeVaultFileName(p.DisplayName, p.ID),
		Filters: []runtime.FileFilter{
			{DisplayName: L("PassDepot (*.pd)", "PassDepot (*.pd)"), Pattern: "*.pd"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	return os.WriteFile(path, data, 0o600)
}

// flushVaultFile записывает текущий vault на диск без git push (для резервной копии).
func (a *App) flushVaultFile(profileID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.profileID != profileID || a.vault == nil || !a.dirty {
		return nil
	}
	blob, err := vaultcore.EncryptVault(a.vault, a.masterPw)
	if err != nil {
		return err
	}
	vp, err := vaultLocalPath(profileID)
	if err != nil {
		return err
	}
	if err := atomicWriteFile(vp, blob); err != nil {
		return err
	}
	a.dirty = false
	a.entryDirty = false
	a.touchActivityLocked()
	return nil
}

// PickLocalVaultImport открывает диалог выбора .pd, валидирует файл и держит его
// до ConfirmLocalVaultImport / CancelLocalVaultImport.
func (a *App) PickLocalVaultImport() (LocalVaultImportPickDTO, error) {
	var zero LocalVaultImportPickDTO
	if a.ctx == nil {
		return zero, errors.New(L("нет контекста приложения", "Application context is not available"))
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: L("Импорт локального хранилища PassDepot", "Import Local PassDepot Vault"),
		Filters: []runtime.FileFilter{
			{DisplayName: L("PassDepot (*.pd)", "PassDepot (*.pd)"), Pattern: "*.pd"},
		},
	})
	if err != nil {
		return zero, err
	}
	if path == "" {
		return zero, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	if err := vaultcore.ValidateContainer(data); err != nil {
		return zero, errors.New(L("выбранный файл не является хранилищем PassDepot", "Selected file is not a PassDepot vault"))
	}
	displayName := strings.TrimSpace(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	if displayName == "" {
		displayName = L("Локальное хранилище", "Local vault")
	}

	a.mu.Lock()
	a.clearPendingLocalVaultLocked()
	a.pendingLocalVault = append([]byte(nil), data...)
	a.mu.Unlock()

	return LocalVaultImportPickDTO{Picked: true, SuggestedName: displayName}, nil
}

// ConfirmLocalVaultImport создаёт локальный профиль из файла, выбранного в PickLocalVaultImport.
func (a *App) ConfirmLocalVaultImport(displayName string) (ProfileDTO, error) {
	var zero ProfileDTO
	name := strings.TrimSpace(displayName)
	if name == "" {
		return zero, errors.New(L("укажите имя профиля", "Enter a profile name"))
	}

	a.mu.Lock()
	data := append([]byte(nil), a.pendingLocalVault...)
	a.mu.Unlock()

	if len(data) == 0 {
		return zero, errors.New(L("сначала выберите файл хранилища", "Pick a vault file first"))
	}
	defer zeroBytes(data)

	if err := vaultcore.ValidateContainer(data); err != nil {
		a.CancelLocalVaultImport()
		return zero, errors.New(L("выбранный файл не является хранилищем PassDepot", "Selected file is not a PassDepot vault"))
	}
	p, err := profile.Add(name, "", "", true)
	if err != nil {
		return zero, mapProfileErr(err)
	}
	target, err := profile.LocalVaultPath(p.ID)
	if err != nil {
		_ = profile.Remove(p.ID)
		return zero, err
	}
	if err := atomicWriteFile(target, data); err != nil {
		_ = profile.Remove(p.ID)
		return zero, err
	}
	a.CancelLocalVaultImport()
	return toProfileDTO(p), nil
}

// CancelLocalVaultImport сбрасывает файл, выбранный в PickLocalVaultImport.
func (a *App) CancelLocalVaultImport() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clearPendingLocalVaultLocked()
}

func (a *App) clearPendingLocalVaultLocked() {
	if a.pendingLocalVault != nil {
		zeroBytes(a.pendingLocalVault)
		a.pendingLocalVault = nil
	}
}
