//go:build windows

package credstore

import (
	"testing"
)

func TestPATRoundtrip(t *testing.T) {
	const id = "test-profile-" + "credstore-pat"
	pat := "ghp_test_token_value"

	if err := DeletePAT(id); err != nil {
		t.Fatal(err)
	}
	if err := SetPAT(id, pat); err != nil {
		t.Fatal(err)
	}
	got, err := GetPAT(id)
	if err != nil || got != pat {
		t.Fatalf("got %q err %v", got, err)
	}
	if err := DeletePAT(id); err != nil {
		t.Fatal(err)
	}
	_, err = GetPAT(id)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestDeleteAllPATs(t *testing.T) {
	const id = "test-profile-credstore-delete-all"
	if err := SetPAT(id, "ghp_delete_all"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteAllPATs(); err != nil {
		t.Fatal(err)
	}
	if _, err := GetPAT(id); err == nil {
		t.Fatal("expected missing after DeleteAllPATs")
	}
}
