package appshell

import (
	"os"
	"testing"

	"passdepot/internal/profile"
)

func TestLocalProfileRoundtripWithoutGit(t *testing.T) {
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())
	a := NewApp()
	a.SetAutoLockMinutes(0)

	p, err := a.CreateProfile("Local", "https://unused.example/repo.git", "main", "unused", true)
	if err != nil {
		t.Fatal(err)
	}
	if !p.LocalOnly || p.HasPAT || p.RepoURL != "" || p.Branch != "" {
		t.Fatalf("unexpected dto: %+v", p)
	}
	if err := a.Login(p.ID, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if err := a.AddFolder("Personal"); err != nil {
		t.Fatal(err)
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}

	path, err := profile.LocalVaultPath(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("local vault: %v", err)
	}

	a.Logout()
	if err := a.Login(p.ID, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	v, err := a.GetVault()
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Folders) != 1 || v.Folders[0].Name != "Personal" {
		t.Fatalf("vault not persisted: %+v", v.Folders)
	}
	if err := a.DeleteProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("local vault should be deleted, stat err: %v", err)
	}
}
