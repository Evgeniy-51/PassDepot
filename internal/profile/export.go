package profile

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/google/uuid"
)

// ExportPublic — JSON для переноса на другой ПК (без PAT и пароля).
type ExportPublic struct {
	SchemaVersion   int    `json:"schemaVersion"`
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	RepoURL         string `json:"repoUrl"`
	Branch          string `json:"branch"`
	VaultPathInRepo string `json:"vaultPathInRepo"`
}

// MarshalExportPublic сериализует профиль для экспорта.
func MarshalExportPublic(p Profile) ([]byte, error) {
	p.Normalize()
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if p.VaultFileName == "" {
		p.VaultFileName = MakeVaultFileName(p.DisplayName, p.ID)
	}
	e := ExportPublic{
		SchemaVersion:   ConfigSchemaVersion,
		ID:              p.ID,
		DisplayName:     p.DisplayName,
		RepoURL:         p.RepoURL,
		Branch:          p.Branch,
		VaultPathInRepo: VaultPathInRepo(p),
	}
	return json.MarshalIndent(e, "", "  ")
}

// UnmarshalImportPublic разбирает экспортированный JSON и возвращает Profile для добавления.
// Если в JSON есть id — сохраняем его (тот же файл в passdepot/ на GitHub).
func UnmarshalImportPublic(data []byte) (Profile, error) {
	var e ExportPublic
	if err := json.Unmarshal(data, &e); err != nil {
		return Profile{}, err
	}
	if e.SchemaVersion != ConfigSchemaVersion {
		return Profile{}, errors.New("profile: unsupported export schemaVersion")
	}
	p := Profile{
		ID:          e.ID,
		DisplayName: e.DisplayName,
		RepoURL:     e.RepoURL,
		Branch:      e.Branch,
	}
	if e.VaultPathInRepo != "" {
		p.VaultFileName = filepath.Base(filepath.FromSlash(e.VaultPathInRepo))
		p.VaultFileName = filepath.Base(p.VaultFileName) // защитно
	}
	p.Normalize()
	if p.DisplayName == "" || p.RepoURL == "" {
		return Profile{}, errors.New("profile: import missing fields")
	}
	if p.Branch == "" {
		p.Branch = DefaultBranch
	}
	if p.VaultFileName == "" {
		p.VaultFileName = MakeVaultFileName(p.DisplayName, p.ID)
	}
	return p, nil
}

// AddImported добавляет профиль из экспорта: если id пуст — новый UUID.
func AddImported(p Profile) (Profile, error) {
	if p.DisplayName == "" || p.RepoURL == "" {
		return Profile{}, errors.New("profile: missing fields")
	}
	p.Normalize()
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.VaultFileName == "" {
		p.VaultFileName = MakeVaultFileName(p.DisplayName, p.ID)
	}
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
			return Profile{}, errors.New("profile: id already exists")
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
