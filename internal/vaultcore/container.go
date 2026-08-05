package vaultcore

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	magic   = "PDVT"
	version = 1
)

// headerLenV1 — фиксированная длина заголовка контейнера v1 (magic … nonce).
const headerLenV1 = len(magic) + 1 + 4 + 4 + 1 + 1 + SaltSize + NonceSize // 55

// ValidateContainer проверяет только формат заголовка .pd, без расшифровки.
func ValidateContainer(blob []byte) error {
	if len(blob) < headerLenV1+aeadOverheadX {
		return fmt.Errorf("%w: short file", ErrInvalidFile)
	}
	if string(blob[0:4]) != magic {
		return fmt.Errorf("%w: bad magic", ErrInvalidFile)
	}
	if blob[4] != version {
		return ErrUnsupportedVersion
	}
	return nil
}

// Encrypt seals plaintext с master password. plaintext — обычно JSON vault.
func Encrypt(plaintext []byte, password []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, errors.New("vaultcore: empty password")
	}
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	mem := uint32(DefaultArgonMemoryKiB)
	t := uint32(DefaultArgonTime)
	p := uint8(DefaultArgonThreads)

	key := argon2.IDKey(password, salt, t, mem, p, uint32(KeySize))
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	header := buildHeaderV1(mem, t, p, salt, nonce)
	ciphertext := aead.Seal(nil, nonce, plaintext, header)

	out := make([]byte, 0, len(header)+len(ciphertext))
	out = append(out, header...)
	out = append(out, ciphertext...)
	return out, nil
}

// Decrypt opens container; при успехе возвращает plaintext (JSON).
func Decrypt(blob []byte, password []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, errors.New("vaultcore: empty password")
	}
	if len(blob) < headerLenV1 {
		return nil, fmt.Errorf("%w: short file", ErrInvalidFile)
	}
	if string(blob[0:4]) != magic {
		return nil, fmt.Errorf("%w: bad magic", ErrInvalidFile)
	}
	if blob[4] != version {
		return nil, ErrUnsupportedVersion
	}

	mem := binary.BigEndian.Uint32(blob[5:9])
	t := binary.BigEndian.Uint32(blob[9:13])
	p := blob[13]
	saltLen := blob[14]
	if saltLen != SaltSize {
		return nil, fmt.Errorf("%w: bad salt length", ErrInvalidFile)
	}
	salt := blob[15 : 15+SaltSize]
	nonce := blob[15+SaltSize : headerLenV1]
	header := blob[:headerLenV1]
	ct := blob[headerLenV1:]
	if len(ct) < aeadOverheadX {
		return nil, fmt.Errorf("%w: short ciphertext", ErrInvalidFile)
	}

	key := argon2.IDKey(password, salt, t, mem, p, uint32(KeySize))
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	plain, err := aead.Open(nil, nonce, ct, header)
	if err != nil {
		return nil, ErrDecrypt
	}
	return plain, nil
}

const aeadOverheadX = 16 // Poly1305 tag

func buildHeaderV1(memKib, time uint32, parallelism uint8, salt, nonce []byte) []byte {
	if len(salt) != SaltSize || len(nonce) != NonceSize {
		panic("vaultcore: bad salt/nonce size")
	}
	h := make([]byte, 0, headerLenV1)
	h = append(h, magic...)
	h = append(h, version)
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, memKib)
	h = append(h, tmp...)
	binary.BigEndian.PutUint32(tmp, time)
	h = append(h, tmp...)
	h = append(h, parallelism)
	h = append(h, SaltSize)
	h = append(h, salt...)
	h = append(h, nonce...)
	if len(h) != headerLenV1 {
		panic("vaultcore: header length mismatch")
	}
	return h
}
