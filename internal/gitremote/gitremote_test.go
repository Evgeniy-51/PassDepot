package gitremote

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthHeaderGit(t *testing.T) {
	h := authHeaderGit("ghp_test")
	if !strings.HasPrefix(h, "Authorization: Basic ") {
		t.Fatalf("bad header: %q", h)
	}
}

func TestGitPathAndVersion(t *testing.T) {
	if _, err := exec.LookPath("git.exe"); err != nil {
		if _, err2 := exec.LookPath("git"); err2 != nil {
			t.Skip("git not in PATH")
		}
	}
	v, err := Version()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(v), "git") {
		t.Fatalf("unexpected: %q", v)
	}
}

func TestCommitAddInFreshRepo(t *testing.T) {
	if _, err := GitPath(); err != nil {
		t.Skip("no git")
	}
	dir := t.TempDir()
	initCmd := exec.Command("git", "init")
	initCmd.Dir = dir
	if err := initCmd.Run(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Add(dir, "f.txt"); err != nil {
		t.Fatal(err)
	}
	if err := Commit(dir, "test commit"); err != nil {
		t.Fatal(err)
	}
}
