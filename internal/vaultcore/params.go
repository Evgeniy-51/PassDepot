package vaultcore

// Параметры Argon2id по умолчанию (цель ~200–500 ms на типичном ПК; при необходимости вынесем в настройки).
const (
	DefaultArgonMemoryKiB = 64 * 1024 // 64 MiB
	DefaultArgonTime      = 3
	DefaultArgonThreads   = 4
	SaltSize              = 16
	KeySize               = 32
	NonceSize             = 24 // XChaCha20-Poly1305
)
