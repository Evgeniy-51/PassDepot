package appshell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"passdepot/internal/gitremote"
	"passdepot/internal/profile"
	"passdepot/internal/vaultcore"
)

const noPATMaster = "correct horse battery staple"

// Профиль со свежим uuid: записи PassDepot/PAT/<id> в Credential Manager нет, GetPAT только читает.
func TestGitProfileWithoutPAT(t *testing.T) {
	if _, err := gitremote.GitPath(); err != nil {
		t.Skip("no git")
	}
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())
	a := NewApp()
	a.SetAutoLockMinutes(0)

	p, err := profile.Add("G", "https://invalid.example/repo.git", "main", false)
	if err != nil {
		t.Fatal(err)
	}

	err = a.Login(p.ID, noPATMaster)
	if err == nil || !strings.HasPrefix(err.Error(), needPATPrefix) {
		t.Fatalf("no clone, no PAT: want need-pat error, got %v", err)
	}

	repoDir, err := profile.LocalRepoDir(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	seedClone(t, repoDir, profile.VaultPathInRepo(p))

	if err := a.Login(p.ID, noPATMaster); err != nil {
		t.Fatalf("local copy, no PAT: %v", err)
	}
	if s := a.GetSession(); !s.Unlocked || s.LastError == "" || s.PendingSync {
		t.Fatalf("unexpected session after offline login: %+v", s)
	}

	if err := a.AddFolder("Offline"); err != nil {
		t.Fatal(err)
	}
	if err := a.Save(); err != nil {
		t.Fatalf("save without PAT: %v", err)
	}
	if s := a.GetSession(); s.Dirty || !s.PendingSync || s.LastError != msgSavedNoPAT() {
		t.Fatalf("unexpected session after save: %+v", s)
	}
	lc, err := gitremote.LocalChanges(repoDir, "main", profile.VaultPathInRepo(p))
	if err != nil || lc == nil || lc.Ahead != 1 {
		t.Fatalf("local commit expected: lc=%+v err=%v", lc, err)
	}

	a.Logout()
	if err := a.Login(p.ID, noPATMaster); err != nil {
		t.Fatal(err)
	}
	if s := a.GetSession(); !s.PendingSync {
		t.Fatalf("unpushed commit must keep pendingSync after re-login: %+v", s)
	}
	v, err := a.GetVault()
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Folders) != 1 || v.Folders[0].Name != "Offline" {
		t.Fatalf("offline save lost: %+v", v.Folders)
	}
}

// seedClone создаёт bare-origin с зашифрованной пустой базой и клонирует его в repoDir.
func seedClone(t *testing.T, repoDir, vaultRel string) {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, "origin.git")
	git(t, root, "init", "--bare", bare)
	seed := filepath.Join(root, "seed")
	if err := gitremote.Clone("", bare, seed); err != nil {
		t.Fatal(err)
	}
	if err := gitremote.EnsureBranch(seed, "main"); err != nil {
		t.Fatal(err)
	}
	blob, err := vaultcore.EncryptVault(emptyVault(), []byte(noPATMaster))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seed, vaultRel), blob, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := gitremote.Add(seed, vaultRel); err != nil {
		t.Fatal(err)
	}
	if err := gitremote.Commit(seed, "seed"); err != nil {
		t.Fatal(err)
	}
	if err := gitremote.Push("", seed, "main"); err != nil {
		t.Fatal(err)
	}
	if err := gitremote.Clone("", bare, repoDir); err != nil {
		t.Fatal(err)
	}
	if err := gitremote.EnsureBranch(repoDir, "main"); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
