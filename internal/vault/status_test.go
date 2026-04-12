package vault_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nicholasgasior/envault/internal/vault"
)

func TestStatus_NotInitialised(t *testing.T) {
	tmp := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(tmp, ".vault"),
		RecipientsFile: filepath.Join(tmp, ".vault", "recipients.txt"),
	}
	_, err := vault.Status(cfg)
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestStatus_EmptyVault(t *testing.T) {
	tmp := t.TempDir()
	cfg := vault.DefaultConfig()
	cfg.VaultDir = filepath.Join(tmp, ".vault")
	cfg.RecipientsFile = filepath.Join(tmp, ".vault", "recipients.txt")

	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	statuses, err := vault.Status(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestStatus_DetectsStaleFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := vault.DefaultConfig()
	cfg.VaultDir = filepath.Join(tmp, ".vault")
	cfg.RecipientsFile = filepath.Join(tmp, ".vault", "recipients.txt")

	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create a fake .age file in vault dir
	encPath := filepath.Join(cfg.VaultDir, ".env.age")
	if err := os.WriteFile(encPath, []byte("fake"), 0600); err != nil {
		t.Fatalf("write enc file: %v", err)
	}

	// Create plain file with a newer modification time
	plainPath := filepath.Join(tmp, ".env")
	if err := os.WriteFile(plainPath, []byte("KEY=val"), 0600); err != nil {
		t.Fatalf("write plain file: %v", err)
	}

	// Ensure plain file is newer
	future := time.Now().Add(10 * time.Second)
	if err := os.Chtimes(plainPath, future, future); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	// Run status from tmp dir so relative plain path resolves
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	statuses, err := vault.Status(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if !statuses[0].Stale {
		t.Error("expected file to be marked stale")
	}
	if statuses[0].InSync {
		t.Error("expected file to not be in sync")
	}
}

func TestStatus_InSync(t *testing.T) {
	tmp := t.TempDir()
	cfg := vault.DefaultConfig()
	cfg.VaultDir = filepath.Join(tmp, ".vault")
	cfg.RecipientsFile = filepath.Join(tmp, ".vault", "recipients.txt")

	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	plainPath := filepath.Join(tmp, ".env")
	if err := os.WriteFile(plainPath, []byte("KEY=val"), 0600); err != nil {
		t.Fatalf("write plain file: %v", err)
	}

	// Encrypted file is newer
	encPath := filepath.Join(cfg.VaultDir, ".env.age")
	if err := os.WriteFile(encPath, []byte("fake"), 0600); err != nil {
		t.Fatalf("write enc file: %v", err)
	}
	future := time.Now().Add(10 * time.Second)
	if err := os.Chtimes(encPath, future, future); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	statuses, err := vault.Status(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Stale {
		t.Error("expected file to not be stale")
	}
	if !statuses[0].InSync {
		t.Error("expected file to be in sync")
	}
}
