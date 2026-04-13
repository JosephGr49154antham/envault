package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"

	"github.com/nicholasgasior/envault/internal/recipients"
	"github.com/nicholasgasior/envault/internal/vault"
)

func setupRecipientsVault(t *testing.T) vault.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init vault: %v", err)
	}
	return cfg
}

func TestListRecipients_Empty(t *testing.T) {
	cfg := setupRecipientsVault(t)

	entries, err := vault.ListRecipients(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestListRecipients_ReturnsAll(t *testing.T) {
	cfg := setupRecipientsVault(t)

	// Add two recipients.
	for i := 0; i < 2; i++ {
		id, err := age.GenerateX25519Identity()
		if err != nil {
			t.Fatalf("generate identity: %v", err)
		}
		if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
			t.Fatalf("add recipient: %v", err)
		}
	}

	entries, err := vault.ListRecipients(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.PublicKey == "" {
			t.Error("expected non-empty PublicKey")
		}
		if e.Label == "" {
			t.Error("expected non-empty Label")
		}
	}
}

func TestListRecipients_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}

	_, err := vault.ListRecipients(cfg)
	if err == nil {
		t.Fatal("expected error for uninitialised vault, got nil")
	}
}

func TestListRecipients_MissingFile(t *testing.T) {
	cfg := setupRecipientsVault(t)
	// Remove the recipients file after init.
	if err := os.Remove(cfg.RecipientsFile); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove recipients file: %v", err)
	}

	entries, err := vault.ListRecipients(cfg)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
