//go:build windows

package credstore

import (
	"testing"

	"github.com/danieljoos/wincred"
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

// Тест работает на отдельном префиксе: DeleteAllPATs стёр бы реальные PAT пользователя.
func TestDeleteAllWithPrefix(t *testing.T) {
	const prefix = "PassDepotTest/PAT/"
	for _, name := range []string{prefix + "a", prefix + "b"} {
		c := wincred.NewGenericCredential(name)
		c.CredentialBlob = []byte("ghp_delete_all")
		if err := c.Write(); err != nil {
			t.Fatal(err)
		}
	}
	if err := deleteAllWithPrefix(prefix); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{prefix + "a", prefix + "b"} {
		if _, err := wincred.GetGenericCredential(name); err == nil {
			t.Fatalf("%s still present after deleteAllWithPrefix", name)
		}
	}
}
