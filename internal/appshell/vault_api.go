package appshell

import (
	"errors"
	"strings"
	"time"

	"passdepot/internal/vaultcore"

	"github.com/google/uuid"
)

// GetVault возвращает текущее хранилище (после входа).
func (a *App) GetVault() (*vaultcore.Vault, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil || a.profileID == "" {
		return nil, errors.New(L("не выполнен вход", "Not signed in"))
	}
	return a.vault, nil
}

// AddFolder добавляет папку.
func (a *App) AddFolder(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	id := uuid.NewString()
	order := len(a.vault.Folders)
	a.vault.Folders = append(a.vault.Folders, vaultcore.Folder{
		ID:    id,
		Name:  name,
		Order: order,
	})
	a.dirty = true
	// entryDirty не трогаем — только папка, без подтверждённой пары ключ/значение
	a.touchActivityLocked()
	return nil
}

// RenameFolder переименовывает папку (id не меняется).
func (a *App) RenameFolder(folderID, name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New(L("имя папки не может быть пустым", "Folder name cannot be empty"))
	}
	for i := range a.vault.Folders {
		if a.vault.Folders[i].ID != folderID {
			continue
		}
		a.vault.Folders[i].Name = name
		a.dirty = true
		a.touchActivityLocked()
		return nil
	}
	return errors.New(L("папка не найдена", "Folder not found"))
}

// DeleteFolder удаляет папку и все описания в ней.
func (a *App) DeleteFolder(folderID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	folders := make([]vaultcore.Folder, 0, len(a.vault.Folders))
	found := false
	for _, f := range a.vault.Folders {
		if f.ID == folderID {
			found = true
			continue
		}
		folders = append(folders, f)
	}
	if !found {
		return errors.New(L("папка не найдена", "Folder not found"))
	}
	descs := make([]vaultcore.Description, 0, len(a.vault.Descriptions))
	for _, d := range a.vault.Descriptions {
		if d.FolderID == folderID {
			continue
		}
		descs = append(descs, d)
	}
	a.vault.Folders = folders
	a.vault.Descriptions = descs
	a.dirty = true
	a.entryDirty = true
	a.touchActivityLocked()
	return nil
}

// AddDescription добавляет запись. Обязательно только название; логин/пароль/заметки опциональны.
func (a *App) AddDescription(folderID, title, key, password, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	if strings.TrimSpace(title) == "" {
		return errors.New(L("название не может быть пустым", "Title cannot be empty"))
	}
	id := uuid.NewString()
	a.vault.Descriptions = append(a.vault.Descriptions, vaultcore.Description{
		ID:        id,
		FolderID:  folderID,
		Title:     title,
		Key:       key,
		Password:  password,
		Value:     value,
		UpdatedAt: time.Now().Unix(),
	})
	a.dirty = true
	a.entryDirty = true
	a.touchActivityLocked()
	return nil
}

// UpdateDescriptionTitle обновляет подпись в списке (дубликаты имён в папке разрешены).
func (a *App) UpdateDescriptionTitle(descriptionID, title string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	if strings.TrimSpace(title) == "" {
		return errors.New(L("название не может быть пустым", "Title cannot be empty"))
	}
	for i := range a.vault.Descriptions {
		if a.vault.Descriptions[i].ID != descriptionID {
			continue
		}
		a.vault.Descriptions[i].Title = title
		a.vault.Descriptions[i].UpdatedAt = time.Now().Unix()
		a.dirty = true
		a.entryDirty = true
		a.touchActivityLocked()
		return nil
	}
	return errors.New(L("описание не найдено", "Entry not found"))
}

// UpdateDescriptionKey обновляет логин (пустой допустим).
func (a *App) UpdateDescriptionKey(descriptionID, key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	for i := range a.vault.Descriptions {
		if a.vault.Descriptions[i].ID != descriptionID {
			continue
		}
		a.vault.Descriptions[i].Key = key
		a.vault.Descriptions[i].UpdatedAt = time.Now().Unix()
		a.dirty = true
		a.entryDirty = true
		a.touchActivityLocked()
		return nil
	}
	return errors.New(L("описание не найдено", "Entry not found"))
}

// UpdateDescriptionPassword обновляет пароль (пустой допустим — старые записи / очистка).
func (a *App) UpdateDescriptionPassword(descriptionID, password string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	for i := range a.vault.Descriptions {
		if a.vault.Descriptions[i].ID != descriptionID {
			continue
		}
		a.vault.Descriptions[i].Password = password
		a.vault.Descriptions[i].UpdatedAt = time.Now().Unix()
		a.dirty = true
		a.entryDirty = true
		a.touchActivityLocked()
		return nil
	}
	return errors.New(L("описание не найдено", "Entry not found"))
}

// UpdateDescriptionValue обновляет заметки (пустые допустимы; текст без Trim).
func (a *App) UpdateDescriptionValue(descriptionID, value string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	for i := range a.vault.Descriptions {
		if a.vault.Descriptions[i].ID != descriptionID {
			continue
		}
		a.vault.Descriptions[i].Value = value
		a.vault.Descriptions[i].UpdatedAt = time.Now().Unix()
		a.dirty = true
		a.entryDirty = true
		a.touchActivityLocked()
		return nil
	}
	return errors.New(L("описание не найдено", "Entry not found"))
}

// DeleteDescription удаляет описание по id.
func (a *App) DeleteDescription(descriptionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.vault == nil {
		return errors.New(L("не выполнен вход", "Not signed in"))
	}
	out := make([]vaultcore.Description, 0, len(a.vault.Descriptions))
	found := false
	for _, d := range a.vault.Descriptions {
		if d.ID == descriptionID {
			found = true
			continue
		}
		out = append(out, d)
	}
	if !found {
		return errors.New(L("описание не найдено", "Entry not found"))
	}
	a.vault.Descriptions = out
	a.dirty = true
	a.entryDirty = true
	a.touchActivityLocked()
	return nil
}
