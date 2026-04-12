package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/nicholasgasior/envault/internal/recipients"
	"github.com/nicholasgasior/envault/internal/vault"
)

func generateIdentity(t *testing.T) *age.X25519Identity {
	t.Helper()
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	return id
}

func TestPushPull_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := vault.Init(tmpDir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	id := generateIdentity(t)

	// Register the recipient.
	if err := recipients.AddRecipient(cfg.RecipientsPath, id.Recipient().String()); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	// Write a plaintext .env file.
	plaintext := "DB_HOST=localhost\nDB_PORT=5432\n"
	if err := os.WriteFile(cfg.PlaintextPath, []byte(plaintext), 0600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := vault.Push(cfg, id); err != nil {
		t.Fatalf("Push: %v", err)
	}

	// Remove plaintext and pull back.
	os.Remove(cfg.PlaintextPath)

	if err := vault.Pull(cfg, id); err != nil {
		t.Fatalf("Pull: %v", err)
	}

	got, err := os.ReadFile(cfg.PlaintextPath)
	if err != nil {
		t.Fatalf("read .env after pull: %v", err)
	}
	if string(got) != plaintext {
		t.Errorf("got %q, want %q", string(got), plaintext)
	}
}

func TestPush_NoRecipients(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := vault.Init(tmpDir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Write a plaintext .env file.
	if err := os.WriteFile(cfg.PlaintextPath, []byte("KEY=val\n"), 0600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	id := generateIdentity(t)
	err = vault.Push(cfg, id)
	if err == nil {
		t.Fatal("expected error when no recipients, got nil")
	}
}

func TestPull_EncryptedNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := vault.DefaultConfig(tmpDir)
	// Point encrypted path to a non-existent file.
	cfg.EncryptedPath = filepath.Join(tmpDir, "missing.age")

	id := generateIdentity(t)
	err := vault.Pull(cfg, id)
	if err == nil {
		t.Fatal("expected error for missing encrypted file, got nil")
	}
}
