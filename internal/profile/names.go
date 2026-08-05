package profile

import (
	"errors"
	"strings"
)

// ErrDuplicateDisplayName — имя профиля уже занято (без учёта регистра).
var ErrDuplicateDisplayName = errors.New("profile: duplicate display name")

func sameDisplayName(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// displayNameTakenUnlocked — true, если имя занято другим профилем (excludeID можно пустым).
func displayNameTakenUnlocked(profiles []Profile, displayName, excludeID string) bool {
	for i := range profiles {
		if excludeID != "" && profiles[i].ID == excludeID {
			continue
		}
		if sameDisplayName(profiles[i].DisplayName, displayName) {
			return true
		}
	}
	return false
}
