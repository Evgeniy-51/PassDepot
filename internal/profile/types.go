package profile

import (
	"errors"
	"strings"
)

// ConfigSchemaVersion — версия файла profiles.json.
const ConfigSchemaVersion = 1

// Config — содержимое profiles.json.
type Config struct {
	SchemaVersion int       `json:"schemaVersion"`
	Profiles      []Profile `json:"profiles"`
}

// Profile — метаданные профиля (без PAT).
type Profile struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	RepoURL     string `json:"repoUrl,omitempty"`
	Branch      string `json:"branch,omitempty"`
	LocalOnly   bool   `json:"localOnly,omitempty"`
	// VaultFileName — имя файла vault в репозитории (например "work-<id>.pd").
	// Пусто => совместимость со старым форматом: "<id>.pd".
	VaultFileName string `json:"vaultFileName,omitempty"`
}

// DefaultBranch если пусто в файле.
const DefaultBranch = "main"

// Normalize заполняет Branch по умолчанию и обрезает пробелы.
func (p *Profile) Normalize() {
	p.DisplayName = strings.TrimSpace(p.DisplayName)
	p.RepoURL = strings.TrimSpace(p.RepoURL)
	p.Branch = strings.TrimSpace(p.Branch)
	if !p.LocalOnly && p.Branch == "" {
		p.Branch = DefaultBranch
	}
}

// Validate проверяет поля перед сохранением.
func (p *Profile) Validate() error {
	p.Normalize()
	if p.ID == "" {
		return errors.New("profile: empty id")
	}
	if p.DisplayName == "" {
		return errors.New("profile: empty display name")
	}
	if p.LocalOnly {
		return nil
	}
	if p.RepoURL == "" {
		return errors.New("profile: empty repo url")
	}
	if !strings.HasPrefix(strings.ToLower(p.RepoURL), "https://") {
		return errors.New("profile: repo url must start with https://")
	}
	return nil
}
