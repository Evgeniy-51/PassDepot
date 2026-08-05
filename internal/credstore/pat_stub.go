//go:build !windows

package credstore

import "errors"

// SetPAT — заглушка вне Windows.
func SetPAT(profileID, pat string) error {
	return errors.New("credstore: only supported on Windows")
}

// GetPAT — заглушка вне Windows.
func GetPAT(profileID string) (string, error) {
	return "", errors.New("credstore: only supported on Windows")
}

// DeletePAT — заглушка вне Windows.
func DeletePAT(profileID string) error {
	return errors.New("credstore: only supported on Windows")
}

// DeleteAllPATs — заглушка вне Windows.
func DeleteAllPATs() error {
	return errors.New("credstore: only supported on Windows")
}
