package vaultcore

import (
	"encoding/json"
)

// MarshalVault сериализует vault в JSON (pretty не нужен — бинарник).
func MarshalVault(v *Vault) ([]byte, error) {
	if v == nil {
		v = &Vault{Version: 1}
	}
	if v.Version == 0 {
		v.Version = 1
	}
	return json.Marshal(v)
}

// UnmarshalVault парсит JSON в Vault.
func UnmarshalVault(data []byte) (*Vault, error) {
	var v Vault
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// EncryptVault шифрует структуру Vault в .pd bytes.
func EncryptVault(v *Vault, password []byte) ([]byte, error) {
	plain, err := MarshalVault(v)
	if err != nil {
		return nil, err
	}
	return Encrypt(plain, password)
}

// DecryptVault расшифровывает и парсит Vault.
func DecryptVault(blob []byte, password []byte) (*Vault, error) {
	plain, err := Decrypt(blob, password)
	if err != nil {
		return nil, err
	}
	return UnmarshalVault(plain)
}
