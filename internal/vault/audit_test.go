package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/vault"
)

func setupAuditVault(t *testing.T) vault.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init vault: %v", err)
	}
	return cfg
}

func TestRecordAndLoadAuditLog(t *testing.T) {
	cfg := setupAuditVault(t)

	if err := vault.RecordEvent(cfg, "push", "alice", "encrypted .env"); err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	if err := vault.RecordEvent(cfg, "pull", "bob", "decrypted .env"); err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}

	log, err := vault.LoadAuditLog(cfg)
	if err != nil {
		t.Fatalf("LoadAuditLog: %v", err)
	}
	if len(log.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(log.Events))
	}
	if log.Events[0].Operation != "push" {
		t.Errorf("expected push, got %s", log.Events[0].Operation)
	}
	if log.Events[1].User != "bob" {
		t.Errorf("expected bob, got %s", log.Events[1].User)
	}
}

func TestLoadAuditLog_Empty(t *testing.T) {
	cfg := setupAuditVault(t)

	log, err := vault.LoadAuditLog(cfg)
	if err != nil {
		t.Fatalf("LoadAuditLog: %v", err)
	}
	if len(log.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(log.Events))
	}
}

func TestRecordEvent_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}

	err := vault.RecordEvent(cfg, "push", "alice", "test")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestAuditLog_Persists(t *testing.T) {
	cfg := setupAuditVault(t)

	for _, op := range []string{"push", "rekey", "rotate"} {
		if err := vault.RecordEvent(cfg, op, "ci", ""); err != nil {
			t.Fatalf("RecordEvent %s: %v", op, err)
		}
	}

	path := filepath.Join(cfg.VaultDir, "audit.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("audit.json not found: %v", err)
	}

	log, err := vault.LoadAuditLog(cfg)
	if err != nil {
		t.Fatalf("LoadAuditLog: %v", err)
	}
	if len(log.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(log.Events))
	}
}
