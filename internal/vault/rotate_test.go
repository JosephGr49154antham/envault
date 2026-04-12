package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/recipients"
	"github.com/nicholasgasior/envault/internal/vault"
)

func TestRotate_CreatesBackupAndRekeys(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      dir + "/.envault",
		RecipientsFile: dir + "/.envault/recipients",
		EncryptedFile:  dir + "/.envault/env.age",
	}

	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Generate an identity and add as recipient.
	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	// Push some data so an encrypted file exists.
	envPath := dir + "/.env"
	if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := vault.Push(cfg, envPath, id); err != nil {
		t.Fatalf("push: %v", err)
	}

	// Rotate.
	backupDir := dir + "/backups"
	if err := vault.Rotate(cfg, vault.RotateOptions{BackupDir: backupDir}); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	// Backup directory should contain exactly one file.
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("read backup dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 backup file, got %d", len(entries))
	}
	if !strings.HasSuffix(entries[0].Name(), "env.age") {
		t.Errorf("backup name should end with env.age, got %s", entries[0].Name())
	}

	// Encrypted file must still exist after rotation.
	if _, err := os.Stat(cfg.EncryptedFile); err != nil {
		t.Errorf("encrypted file missing after rotate: %v", err)
	}
}

func TestRotate_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      dir + "/.envault",
		RecipientsFile: dir + "/.envault/recipients",
		EncryptedFile:  dir + "/.envault/env.age",
	}

	err := vault.Rotate(cfg, vault.RotateOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestRotate_DefaultBackupDir(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      dir + "/.envault",
		RecipientsFile: dir + "/.envault/recipients",
		EncryptedFile:  dir + "/.envault/env.age",
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	envPath := dir + "/.env"
	_ = os.WriteFile(envPath, []byte("A=1\n"), 0o600)
	_ = vault.Push(cfg, envPath, id)

	if err := vault.Rotate(cfg, vault.RotateOptions{}); err != nil {
		t.Fatalf("rotate with default backup dir: %v", err)
	}

	defaultBackup := filepath.Join(cfg.VaultDir, "backups")
	if _, err := os.Stat(defaultBackup); err != nil {
		t.Errorf("default backup dir not created: %v", err)
	}
}
