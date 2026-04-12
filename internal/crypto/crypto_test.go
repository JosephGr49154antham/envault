package crypto_test

import (
	"testing"

	"filippo.io/age"

	"github.com/yourusername/envault/internal/crypto"
)

func generateTestRecipientIdentity(t *testing.T) (*age.X25519Identity, *age.X25519Recipient) {
	t.Helper()
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generating identity: %v", err)
	}
	return id, id.Recipient()
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	id, recipient := generateTestRecipientIdentity(t)

	plaintext := []byte("SECRET_KEY=supersecret\nDB_PASS=hunter2")

	ciphertext, err := crypto.Encrypt(plaintext, []age.Recipient{recipient})
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("expected non-empty ciphertext")
	}

	got, err := crypto.Decrypt(ciphertext, []age.Identity{id})
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if string(got) != string(plaintext) {
		t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestDecrypt_WrongIdentity(t *testing.T) {
	_, recipient := generateTestRecipientIdentity(t)
	wrongID, _ := generateTestRecipientIdentity(t)

	plaintext := []byte("TOP_SECRET=yes")
	ciphertext, err := crypto.Encrypt(plaintext, []age.Recipient{recipient})
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = crypto.Decrypt(ciphertext, []age.Identity{wrongID})
	if err == nil {
		t.Fatal("expected error decrypting with wrong identity, got nil")
	}
}

func TestEncrypt_EmptyPlaintext(t *testing.T) {
	_, recipient := generateTestRecipientIdentity(t)

	ciphertext, err := crypto.Encrypt([]byte{}, []age.Recipient{recipient})
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("expected non-empty ciphertext even for empty input")
	}
}
