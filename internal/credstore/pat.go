//go:build windows

package credstore

import (
	"errors"
	"strings"

	"github.com/danieljoos/wincred"
)

const targetPrefix = "PassDepot/PAT/"

func targetName(profileID string) string {
	return targetPrefix + profileID
}

// SetPAT сохраняет GitHub PAT в Windows Credential Manager.
func SetPAT(profileID, pat string) error {
	profileID = strings.TrimSpace(profileID)
	pat = strings.TrimSpace(pat)
	if profileID == "" {
		return errors.New("credstore: empty profile id")
	}
	if pat == "" {
		return errors.New("credstore: empty pat")
	}
	c := wincred.NewGenericCredential(targetName(profileID))
	c.CredentialBlob = []byte(pat)
	return c.Write()
}

// GetPAT возвращает PAT или ошибку, если не найден.
func GetPAT(profileID string) (string, error) {
	if strings.TrimSpace(profileID) == "" {
		return "", errors.New("credstore: empty profile id")
	}
	c, err := wincred.GetGenericCredential(targetName(profileID))
	if err != nil {
		return "", err
	}
	return string(c.CredentialBlob), nil
}

// DeletePAT удаляет сохранённый PAT. Если записи нет — ErrNotExist не обязателен (игнорируем).
func DeletePAT(profileID string) error {
	if strings.TrimSpace(profileID) == "" {
		return errors.New("credstore: empty profile id")
	}
	c, err := wincred.GetGenericCredential(targetName(profileID))
	if err != nil {
		return nil // уже удалено
	}
	return c.Delete()
}

// DeleteAllPATs удаляет все записи PassDepot/PAT/* из Credential Manager.
func DeleteAllPATs() error {
	list, err := wincred.FilteredList(targetPrefix + "*")
	if err != nil {
		return nil
	}
	var first error
	for _, c := range list {
		if c == nil || c.TargetName == "" {
			continue
		}
		g, err := wincred.GetGenericCredential(c.TargetName)
		if err != nil {
			continue
		}
		if err := g.Delete(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
