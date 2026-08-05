package vaultcore

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	pw := []byte("correct horse battery staple")
	v := &Vault{
		Version: 1,
		Folders: []Folder{{ID: "f1", Name: "Work", Order: 0}},
		Descriptions: []Description{{
			ID: "d1", FolderID: "f1", Title: "GitHub", Key: "user", Password: "secret", Value: "note", UpdatedAt: 1,
		}},
	}

	blob, err := EncryptVault(v, pw)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecryptVault(blob, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Folders) != 1 || out.Folders[0].Name != "Work" {
		t.Fatalf("folders: %+v", out.Folders)
	}
	if len(out.Descriptions) != 1 || out.Descriptions[0].Password != "secret" || out.Descriptions[0].Value != "note" {
		t.Fatalf("descriptions: %+v", out.Descriptions)
	}
}

func TestUnmarshalLegacyDescriptionWithoutPassword(t *testing.T) {
	raw := []byte(`{"version":1,"folders":[{"id":"f1","name":"Work","order":0}],"descriptions":[{"id":"d1","folderId":"f1","title":"GitHub","key":"site","value":"old-secret","updatedAt":1}]}`)
	v, err := UnmarshalVault(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Descriptions) != 1 {
		t.Fatalf("want 1 desc, got %+v", v.Descriptions)
	}
	d := v.Descriptions[0]
	if d.Title != "GitHub" || d.Key != "site" || d.Value != "old-secret" {
		t.Fatalf("legacy fields: %+v", d)
	}
	if d.Password != "" {
		t.Fatalf("password should default empty, got %q", d.Password)
	}
}

func TestWrongPassword(t *testing.T) {
	blob, err := Encrypt([]byte("hello"), []byte("good"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decrypt(blob, []byte("wrong"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrDecrypt) {
		t.Fatalf("got %v want %v", err, ErrDecrypt)
	}
}

func TestCorruptedTruncated(t *testing.T) {
	blob, err := Encrypt([]byte("x"), []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	short := blob[:headerLenV1+5]
	_, err = Decrypt(short, []byte("pw"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidFile) && !errors.Is(err, ErrDecrypt) {
		t.Fatalf("got %v", err)
	}
}

func TestBadMagic(t *testing.T) {
	blob, err := Encrypt([]byte("a"), []byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	blob[0] = 'X'
	_, err = Decrypt(blob, []byte("b"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidFile) {
		t.Fatalf("got %v", err)
	}
}

func TestTamperedCiphertext(t *testing.T) {
	blob, err := Encrypt([]byte("payload"), []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	if len(blob) < headerLenV1+1 {
		t.Fatal("short blob")
	}
	blob[len(blob)-1] ^= 0xFF
	_, err = Decrypt(blob, []byte("pw"))
	if !errors.Is(err, ErrDecrypt) {
		t.Fatalf("got %v", err)
	}
}

func TestEmptyPasswordEncrypt(t *testing.T) {
	_, err := Encrypt([]byte("x"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = Encrypt([]byte("x"), []byte{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEncryptDeterministicHeaderLayout(t *testing.T) {
	b1, _ := Encrypt([]byte("same"), []byte("pw"))
	b2, _ := Encrypt([]byte("same"), []byte("pw"))
	if bytes.Equal(b1, b2) {
		t.Fatal("nonce/salt must differ")
	}
	if len(b1) != len(b2) {
		// ciphertext same length for same plaintext
		t.Fatalf("len %d vs %d", len(b1), len(b2))
	}
}
