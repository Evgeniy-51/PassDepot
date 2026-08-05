package appshell

import (
	"errors"

	"passdepot/internal/profile"
)

func mapProfileErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, profile.ErrDuplicateDisplayName) {
		return errors.New(L("Профиль с таким именем уже есть", "A profile with this name already exists"))
	}
	return err
}
