package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

var storeMu sync.RWMutex

// Load читает profiles.json; если файла нет — пустой конфиг.
func Load() (*Config, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	path, err := ProfilesPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{SchemaVersion: ConfigSchemaVersion, Profiles: nil}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.SchemaVersion == 0 {
		c.SchemaVersion = ConfigSchemaVersion
	}
	return &c, nil
}

// Save атомарно перезаписывает profiles.json.
func Save(c *Config) error {
	if c == nil {
		return errors.New("profile: nil config")
	}
	c.SchemaVersion = ConfigSchemaVersion

	storeMu.Lock()
	defer storeMu.Unlock()

	if err := EnsureDataRoot(); err != nil {
		return err
	}
	path, err := ProfilesPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "profiles-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

// List возвращает копию списка профилей.
func List() ([]Profile, error) {
	c, err := Load()
	if err != nil {
		return nil, err
	}
	out := make([]Profile, len(c.Profiles))
	copy(out, c.Profiles)
	return out, nil
}

// Get по id.
func Get(id string) (Profile, bool) {
	c, err := Load()
	if err != nil {
		return Profile{}, false
	}
	for i := range c.Profiles {
		if c.Profiles[i].ID == id {
			p := c.Profiles[i]
			p.Normalize()
			return p, true
		}
	}
	return Profile{}, false
}

// Add добавляет профиль с новым UUID.
func Add(displayName, repoURL, branch string, localOnly bool) (Profile, error) {
	p := Profile{
		ID:          uuid.NewString(),
		DisplayName: displayName,
		RepoURL:     repoURL,
		Branch:      branch,
		LocalOnly:   localOnly,
	}
	p.VaultFileName = MakeVaultFileName(p.DisplayName, p.ID)
	p.Normalize()
	if err := p.Validate(); err != nil {
		return Profile{}, err
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	c, err := loadUnlocked()
	if err != nil {
		return Profile{}, err
	}
	for i := range c.Profiles {
		if c.Profiles[i].ID == p.ID {
			return Profile{}, errors.New("profile: id collision")
		}
	}
	if displayNameTakenUnlocked(c.Profiles, p.DisplayName, "") {
		return Profile{}, ErrDuplicateDisplayName
	}
	c.Profiles = append(c.Profiles, p)
	if err := saveUnlocked(c); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// Update обновляет поля существующего профиля.
func Update(p Profile) error {
	if err := p.Validate(); err != nil {
		return err
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	c, err := loadUnlocked()
	if err != nil {
		return err
	}
	for i := range c.Profiles {
		if c.Profiles[i].ID == p.ID {
			if displayNameTakenUnlocked(c.Profiles, p.DisplayName, p.ID) {
				return ErrDuplicateDisplayName
			}
			c.Profiles[i] = p
			return saveUnlocked(c)
		}
	}
	return errors.New("profile: not found")
}

// Remove удаляет профиль из конфига (PAT и каталог repos — отдельно).
func Remove(id string) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	c, err := loadUnlocked()
	if err != nil {
		return err
	}
	n := c.Profiles[:0]
	for _, p := range c.Profiles {
		if p.ID != id {
			n = append(n, p)
		}
	}
	if len(n) == len(c.Profiles) {
		return errors.New("profile: not found")
	}
	c.Profiles = n
	return saveUnlocked(c)
}

func loadUnlocked() (*Config, error) {
	path, err := ProfilesPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{SchemaVersion: ConfigSchemaVersion, Profiles: nil}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.SchemaVersion == 0 {
		c.SchemaVersion = ConfigSchemaVersion
	}
	return &c, nil
}

func saveUnlocked(c *Config) error {
	if err := EnsureDataRoot(); err != nil {
		return err
	}
	path, err := ProfilesPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "profiles-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}
