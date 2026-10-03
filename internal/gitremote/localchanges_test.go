package gitremote

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const testVault = "v.pd"

func needGit(t *testing.T) {
	t.Helper()
	if _, err := GitPath(); err != nil {
		t.Skip("no git")
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeVault(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, testVault), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readVault(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, testVault))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func commitVault(t *testing.T, dir, content string) {
	t.Helper()
	writeVault(t, dir, content)
	if err := Add(dir, testVault); err != nil {
		t.Fatal(err)
	}
	if err := Commit(dir, "c "+content); err != nil {
		t.Fatal(err)
	}
}

// newRemote создаёт bare origin с одним коммитом в main и возвращает его путь.
func newRemote(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, "origin.git")
	gitRun(t, root, "init", "--bare", bare)
	seed := filepath.Join(root, "seed")
	if err := Clone("", bare, seed); err != nil {
		t.Fatal(err)
	}
	if err := EnsureBranch(seed, "main"); err != nil {
		t.Fatal(err)
	}
	commitVault(t, seed, "v1")
	if err := Push("", seed, "main"); err != nil {
		t.Fatal(err)
	}
	return bare
}

func cloneTo(t *testing.T, bare string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "w")
	if err := Clone("", bare, dir); err != nil {
		t.Fatal(err)
	}
	if err := EnsureBranch(dir, "main"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func asLocalChanges(t *testing.T, err error) *LocalChangesError {
	t.Helper()
	var lc *LocalChangesError
	if !errors.As(err, &lc) {
		t.Fatalf("want *LocalChangesError, got %v", err)
	}
	return lc
}

func TestRefreshCleanResetsToOrigin(t *testing.T) {
	needGit(t)
	bare := newRemote(t)
	a := cloneTo(t, bare)
	b := cloneTo(t, bare)

	commitVault(t, a, "v2")
	if err := Push("", a, "main"); err != nil {
		t.Fatal(err)
	}
	if err := Refresh("", b, "main", testVault); err != nil {
		t.Fatal(err)
	}
	if got := readVault(t, b); got != "v2" {
		t.Fatalf("vault = %q, want v2", got)
	}
}

func TestRefreshKeepsUnpushedCommit(t *testing.T) {
	needGit(t)
	bare := newRemote(t)
	b := cloneTo(t, bare)

	commitVault(t, b, "local")
	lc := asLocalChanges(t, Refresh("", b, "main", testVault))
	if lc.Ahead != 1 || lc.Behind != 0 || lc.Uncommitted {
		t.Fatalf("unexpected %+v", lc)
	}
	if got := readVault(t, b); got != "local" {
		t.Fatalf("unpushed commit lost: vault = %q", got)
	}
}

func TestRefreshKeepsUncommittedVault(t *testing.T) {
	needGit(t)
	bare := newRemote(t)
	b := cloneTo(t, bare)

	writeVault(t, b, "dirty")
	lc := asLocalChanges(t, Refresh("", b, "main", testVault))
	if lc.Ahead != 0 || !lc.Uncommitted {
		t.Fatalf("unexpected %+v", lc)
	}
	if got := readVault(t, b); got != "dirty" {
		t.Fatalf("uncommitted vault lost: vault = %q", got)
	}
}

func TestRefreshIgnoresOtherUncommittedFiles(t *testing.T) {
	needGit(t)
	bare := newRemote(t)
	b := cloneTo(t, bare)

	if err := os.WriteFile(filepath.Join(b, "other.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Refresh("", b, "main", testVault); err != nil {
		t.Fatalf("untracked non-vault file must not block refresh: %v", err)
	}
}

func TestRefreshReportsDivergence(t *testing.T) {
	needGit(t)
	bare := newRemote(t)
	a := cloneTo(t, bare)
	b := cloneTo(t, bare)

	commitVault(t, a, "remote")
	if err := Push("", a, "main"); err != nil {
		t.Fatal(err)
	}
	commitVault(t, b, "local")
	lc := asLocalChanges(t, Refresh("", b, "main", testVault))
	if lc.Ahead != 1 || lc.Behind != 1 {
		t.Fatalf("unexpected %+v", lc)
	}
	if got := readVault(t, b); got != "local" {
		t.Fatalf("local commit lost: vault = %q", got)
	}
}

func TestLocalChangesEmptyRemote(t *testing.T) {
	needGit(t)
	root := t.TempDir()
	bare := filepath.Join(root, "origin.git")
	gitRun(t, root, "init", "--bare", bare)
	dir := filepath.Join(root, "w")
	if err := Clone("", bare, dir); err != nil {
		t.Fatal(err)
	}
	if err := EnsureBranch(dir, "main"); err != nil {
		t.Fatal(err)
	}

	lc, err := LocalChanges(dir, "main", testVault)
	if err != nil || lc != nil {
		t.Fatalf("empty clone: lc=%+v err=%v", lc, err)
	}

	commitVault(t, dir, "first")
	lc, err = LocalChanges(dir, "main", testVault)
	if err != nil {
		t.Fatal(err)
	}
	if lc == nil || lc.Ahead != 1 {
		t.Fatalf("unpushed first commit not detected: %+v", lc)
	}
}

func TestLocalChangesClean(t *testing.T) {
	needGit(t)
	bare := newRemote(t)
	b := cloneTo(t, bare)
	lc, err := LocalChanges(b, "main", testVault)
	if err != nil || lc != nil {
		t.Fatalf("clean clone: lc=%+v err=%v", lc, err)
	}
}
