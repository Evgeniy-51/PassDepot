package vaultcore

import "errors"

var (
	// ErrInvalidFile — слишком короткий файл, неверный magic или заголовок.
	ErrInvalidFile = errors.New("vaultcore: invalid vault file")
	// ErrUnsupportedVersion — версия контейнера не поддерживается.
	ErrUnsupportedVersion = errors.New("vaultcore: unsupported container version")
	// ErrDecrypt — не удалось расшифровать (неверный пароль или подделка ciphertext).
	ErrDecrypt = errors.New("vaultcore: decrypt failed (wrong password or tampered data)")
)
