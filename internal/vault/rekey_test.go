package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"

	"github.com/envault/envault/internal/vault"
)

func generateRekeyIdentity(t *testing.T) *age.X25519Identity {
	t.Helper()
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	return id
}

func TestRekey_UpdatesEncryptedFile(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.DefaultConfig(dir)

	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	oldID := generateRekeyIdentity(t)
	newID := generateRekeyIdentity(t)

	// Write a plaintext .env file
	envContent := []byte("SECRET=hello\nOTHER=world\n")
	if err := os.WriteFile(cfg.EnvFile, envContent, 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	// Write old recipient
	recFile := cfg.RecipientsFile
	if err := os.WriteFile(recFile, []byte(oldID.Recipient().String()+"\n"), 0644); err != nil {
		t.Fatalf("write recipients: %v", err)
	}

	// Push with old identity
	if err := vault.Push(cfg, oldID); err != nil {
		t.Fatalf("push: %v", err)
	}

	// Replace recipient with new identity
	if err := os.WriteFile(recFile, []byte(newID.Recipient().String()+"\n"), 0644); err != nil {
		t.Fatalf("update recipients: %v", err)
	}

	// Rekey: decrypt with old, re-encrypt with new
	if err := vault.Rekey(cfg, oldID); err != nil {
		t.Fatalf("rekey: %v", err)
	}

	// New identity should be able to pull
	dst := filepath.Join(dir, "decrypted.env")
	if err := vault.Pull(cfg, newID, dst); err != nil {
		t.Fatalf("pull after rekey: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read decrypted: %v", err)
	}
	if string(got) != string(envContent) {
		t.Errorf("content mismatch: got %q, want %q", got, envContent)
	}
}

func TestRekey_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.DefaultConfig(dir)
	id := generateRekeyIdentity(t)

	if err := vault.Rekey(cfg, id); err == nil {
		t.Error("expected error for uninitialised vault, got nil")
	}
}

func TestRekey_NoEncryptedFile(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.DefaultConfig(dir)

	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	id := generateRekeyIdentity(t)
	// No encrypted file present — Rekey should return an error
	if err := vault.Rekey(cfg, id); err == nil {
		t.Error("expected error when encrypted file missing, got nil")
	}
}
