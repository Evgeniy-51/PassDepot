package appshell

import (
	"os"
	"testing"

	"passdepot/internal/profile"
	"passdepot/internal/vaultcore"
)

func TestConfirmLocalVaultImportUsesCustomName(t *testing.T) {
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())
	a := NewApp()

	blob, err := vaultcore.EncryptVault(&vaultcore.Vault{Version: 1}, []byte("import-master-pw"))
	if err != nil {
		t.Fatal(err)
	}
	a.mu.Lock()
	a.pendingLocalVault = append([]byte(nil), blob...)
	a.mu.Unlock()

	if _, err := a.ConfirmLocalVaultImport("   "); err == nil {
		t.Fatal("expected empty name error")
	}
	a.mu.Lock()
	stillPending := len(a.pendingLocalVault) > 0
	a.mu.Unlock()
	if !stillPending {
		t.Fatal("pending vault should remain after empty name")
	}

	p, err := a.ConfirmLocalVaultImport("My Imported Vault")
	if err != nil {
		t.Fatal(err)
	}
	if p.DisplayName != "My Imported Vault" || !p.LocalOnly {
		t.Fatalf("unexpected dto: %+v", p)
	}
	path, err := profile.LocalVaultPath(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("vault file: %v", err)
	}
	a.mu.Lock()
	cleared := a.pendingLocalVault == nil
	a.mu.Unlock()
	if !cleared {
		t.Fatal("pending vault should be cleared after confirm")
	}

	if _, err := a.ConfirmLocalVaultImport("Again"); err == nil {
		t.Fatal("expected error when nothing pending")
	}
}

func TestCancelLocalVaultImport(t *testing.T) {
	a := NewApp()
	a.mu.Lock()
	a.pendingLocalVault = []byte{1, 2, 3}
	a.mu.Unlock()
	a.CancelLocalVaultImport()
	a.mu.Lock()
	cleared := a.pendingLocalVault == nil
	a.mu.Unlock()
	if !cleared {
		t.Fatal("pending not cleared")
	}
}
