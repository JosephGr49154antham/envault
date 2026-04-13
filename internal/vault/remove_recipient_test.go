package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/vault"
)

func setupRemoveVault(t *testing.T) (vault.Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedFile:  filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		IdentityFile:   filepath.Join(dir, ".envault", "identity.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Generate and save an identity so Rekey can load it.
	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	if err := keymgr.SaveIdentity(id, cfg.IdentityFile); err != nil {
		t.Fatalf("save identity: %v", err)
	}

	return cfg, id.Recipient().String()
}

func TestRemoveRecipient_RemovesKey(t *testing.T) {
	cfg, ownerKey := setupRemoveVault(t)

	// Write two recipients.
	extraID, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate extra key: %v", err)
	}
	extraKey := extraID.Recipient().String()

	content := ownerKey + "\n" + extraKey + "\n"
	if err := os.WriteFile(cfg.RecipientsFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write recipients: %v", err)
	}

	if err := vault.RemoveRecipient(cfg, extraKey); err != nil {
		t.Fatalf("RemoveRecipient: %v", err)
	}

	data, err := os.ReadFile(cfg.RecipientsFile)
	if err != nil {
		t.Fatalf("read recipients: %v", err)
	}
	if strings.Contains(string(data), extraKey) {
		t.Error("expected extra key to be removed but it is still present")
	}
	if !strings.Contains(string(data), ownerKey) {
		t.Error("expected owner key to remain but it was removed")
	}
}

func TestRemoveRecipient_NotFound(t *testing.T) {
	cfg, ownerKey := setupRemoveVault(t)

	if err := os.WriteFile(cfg.RecipientsFile, []byte(ownerKey+"\n"), 0o644); err != nil {
		t.Fatalf("write recipients: %v", err)
	}

	err := vault.RemoveRecipient(cfg, "age1nonexistentkey")
	if err == nil {
		t.Fatal("expected error for missing recipient, got nil")
	}
	if !strings.Contains(err.Error(), "recipient not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRemoveRecipient_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedFile:  filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		IdentityFile:   filepath.Join(dir, ".envault", "identity.age"),
	}

	err := vault.RemoveRecipient(cfg, "age1somekey")
	if err == nil {
		t.Fatal("expected error for uninitialised vault, got nil")
	}
	if !strings.Contains(err.Error(), "not initialised") {
		t.Errorf("unexpected error: %v", err)
	}
}
