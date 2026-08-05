package appshell

import (
	"os"
	"path/filepath"

	"passdepot/internal/credstore"
	"passdepot/internal/profile"
)

// RunUninstallCleanup удаляет данные приложения и все PAT из Credential Manager.
// Вызывается uninstall'ом: PassDepot.exe --uninstall-cleanup (до удаления бинарника).
func RunUninstallCleanup() error {
	_ = credstore.DeleteAllPATs()

	if root, err := profile.DataRoot(); err == nil && root != "" {
		_ = os.RemoveAll(root)
	}

	// WebView2 user-data (исторический путь шаблона Wails NSIS).
	if cfg, err := os.UserConfigDir(); err == nil && cfg != "" {
		_ = os.RemoveAll(filepath.Join(cfg, "PassDepot.exe"))
	}
	return nil
}
