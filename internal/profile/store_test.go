package profile

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestDataRootEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PASSDEPOT_DATA_ROOT", dir)
	root, err := DataRoot()
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	if root != abs {
		t.Fatalf("root %q want %q", root, abs)
	}
}

func TestAddListRemove(t *testing.T) {
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())

	p, err := Add("Test", "https://github.com/x/y.git", "main", false)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID == "" {
		t.Fatal("empty id")
	}
	if VaultPathInRepo(p) == p.ID+".pd" {
		t.Fatal("expected prefixed vault filename")
	}

	list, err := List()
	if err != nil || len(list) != 1 {
		t.Fatalf("list %+v err %v", list, err)
	}

	repo, err := LocalRepoDir(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(repo) != p.ID {
		t.Fatal(repo)
	}

	if err := Remove(p.ID); err != nil {
		t.Fatal(err)
	}
	list, _ = List()
	if len(list) != 0 {
		t.Fatal("expected empty")
	}
}

func TestMakeExportProfileFileName(t *testing.T) {
	if got := MakeExportProfileFileName("Work Git"); got != "Work-Git-passdepot-profile.json" {
		t.Fatalf("got %q", got)
	}
	if got := MakeExportProfileFileName(""); got != "passdepot-profile.json" {
		t.Fatalf("empty: got %q", got)
	}
	if got := MakeExportProfileFileName("a/b:c"); got != "a-b-c-passdepot-profile.json" {
		t.Fatalf("sanitize: got %q", got)
	}
}

func TestValidateHTTPS(t *testing.T) {
	p := Profile{ID: "x", DisplayName: "a", RepoURL: "http://x", Branch: "main"}
	if err := p.Validate(); err == nil {
		t.Fatal("want error for non-https")
	}
}

func TestExportImport(t *testing.T) {
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())

	p, err := Add("E", "https://github.com/a/b.git", "", false)
	if err != nil {
		t.Fatal(err)
	}
	data, err := MarshalExportPublic(p)
	if err != nil {
		t.Fatal(err)
	}
	ip, err := UnmarshalImportPublic(data)
	if err != nil {
		t.Fatal(err)
	}
	if ip.ID != p.ID {
		t.Fatalf("id %q %q", ip.ID, p.ID)
	}

	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())
	p2, err := AddImported(ip)
	if err != nil {
		t.Fatal(err)
	}
	if p2.ID != p.ID {
		t.Fatal("id should match export")
	}
}

func TestLocalProfileNeedsNoRepository(t *testing.T) {
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())
	p, err := Add("Local", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if !p.LocalOnly || p.RepoURL != "" || p.Branch != "" {
		t.Fatalf("unexpected local profile: %+v", p)
	}
}

func TestDuplicateDisplayNameRejected(t *testing.T) {
	t.Setenv("PASSDEPOT_DATA_ROOT", t.TempDir())
	if _, err := Add("Home", "", "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := Add("home", "", "", true); !errors.Is(err, ErrDuplicateDisplayName) {
		t.Fatalf("want ErrDuplicateDisplayName, got %v", err)
	}
	if _, err := Add("  HOME  ", "", "", true); !errors.Is(err, ErrDuplicateDisplayName) {
		t.Fatalf("want ErrDuplicateDisplayName for trimmed, got %v", err)
	}

	p2, err := Add("Work", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	p2.DisplayName = "HOME"
	if err := Update(p2); !errors.Is(err, ErrDuplicateDisplayName) {
		t.Fatalf("rename to taken name: %v", err)
	}
	p2.DisplayName = "work" // same profile, case-only change — ok
	if err := Update(p2); err != nil {
		t.Fatal(err)
	}

	imp := Profile{DisplayName: "Home", RepoURL: "https://github.com/a/b.git", Branch: "main"}
	if _, err := AddImported(imp); !errors.Is(err, ErrDuplicateDisplayName) {
		t.Fatalf("import duplicate: %v", err)
	}
}
