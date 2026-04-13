package vault

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"

	"github.com/cqroot/envault/internal/crypto"
	"github.com/cqroot/envault/internal/recipients"
)

func setupVerifyVault(t *testing.T) (Config, *age.X25519Identity) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		IdentityFile:  filepath.Join(dir, "identity.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}

	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}

	// Write identity file.
	if err := os.WriteFile(cfg.IdentityFile, []byte(id.String()+"\n"), 0600); err != nil {
		t.Fatalf("write identity: %v", err)
	}

	// Register recipient.
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	return cfg, id
}

func TestVerify_ValidFiles(t *testing.T) {
	cfg, id := setupVerifyVault(t)

	plainPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(plainPath, []byte("KEY=value\nFOO=bar\n"), 0644); err != nil {
		t.Fatal(err)
	}

	encPath := filepath.Join(cfg.VaultDir, ".env.age")
	if err := crypto.EncryptFile(plainPath, encPath, []age.Recipient{id.Recipient()}); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	results, err := Verify(cfg)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Valid {
		t.Errorf("expected valid, got error: %v", results[0].Error)
	}
	if results[0].Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestVerify_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		IdentityFile:  filepath.Join(dir, "identity.txt"),
	}
	_, err := Verify(cfg)
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestVerify_EmptyVault(t *testing.T) {
	cfg, _ := setupVerifyVault(t)

	results, err := Verify(cfg)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
