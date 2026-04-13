package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"

	"github.com/envault/envault/internal/recipients"
	"github.com/envault/envault/internal/vault"
)

func setupRemoveRekeyVault(t *testing.T) (vault.Config, *age.X25519Identity) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
		PlainFile:     filepath.Join(dir, ".env"),
		IdentityFile:  filepath.Join(dir, ".envault", "identity.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}

	// Write a plain .env file and push so there is something to rekey.
	if err := os.WriteFile(cfg.PlainFile, []byte("SECRET=hello\n"), 0600); err != nil {
		t.Fatalf("write plain: %v", err)
	}

	// Add the identity as a recipient so Push succeeds.
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	if err := vault.Push(cfg); err != nil {
		t.Fatalf("push: %v", err)
	}

	return cfg, id
}

func TestRemoveRecipientAndRekey_RemovesAndRekeys(t *testing.T) {
	cfg, id := setupRemoveRekeyVault(t)

	// Add a second recipient to keep at least one after removal.
	extraID, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate extra identity: %v", err)
	}
	if err := recipients.AddRecipient(cfg.RecipientsFile, extraID.Recipient().String()); err != nil {
		t.Fatalf("add extra recipient: %v", err)
	}

	if err := vault.RemoveRecipientAndRekey(cfg, id.Recipient().String()); err != nil {
		t.Fatalf("RemoveRecipientAndRekey: %v", err)
	}

	// The removed key must no longer appear in the recipients file.
	loaded, err := recipients.LoadRecipients(cfg.RecipientsFile)
	if err != nil {
		t.Fatalf("load recipients: %v", err)
	}
	for _, r := range loaded {
		if r.String() == id.Recipient().String() {
			t.Error("removed recipient still present in recipients file")
		}
	}

	// The encrypted file must still exist (rekey succeeded).
	if _, err := os.Stat(cfg.EncryptedFile); os.IsNotExist(err) {
		t.Error("encrypted file missing after rekey")
	}
}

func TestRemoveRecipientAndRekey_NotFound(t *testing.T) {
	cfg, _ := setupRemoveRekeyVault(t)

	err := vault.RemoveRecipientAndRekey(cfg, "age1notarealkey")
	if err == nil {
		t.Error("expected error when removing non-existent recipient")
	}
}

func TestRemoveRecipientAndRekey_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
		PlainFile:     filepath.Join(dir, ".env"),
		IdentityFile:  filepath.Join(dir, ".envault", "identity.age"),
	}

	err := vault.RemoveRecipientAndRekey(cfg, "age1somekey")
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}
