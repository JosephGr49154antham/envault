package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roamingthings/envault/internal/vault"
)

func setupLockVault(t *testing.T) vault.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg
}

func TestLock_CreatesLockFile(t *testing.T) {
	cfg := setupLockVault(t)

	if err := vault.Lock(cfg); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if !vault.IsLocked(cfg) {
		t.Fatal("expected vault to be locked")
	}
}

func TestLock_FailsWhenAlreadyLocked(t *testing.T) {
	cfg := setupLockVault(t)

	if err := vault.Lock(cfg); err != nil {
		t.Fatalf("first Lock: %v", err)
	}
	if err := vault.Lock(cfg); err == nil {
		t.Fatal("expected error on second Lock, got nil")
	}
}

func TestUnlock_RemovesLockFile(t *testing.T) {
	cfg := setupLockVault(t)

	_ = vault.Lock(cfg)
	if err := vault.Unlock(cfg); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if vault.IsLocked(cfg) {
		t.Fatal("expected vault to be unlocked")
	}
}

func TestUnlock_FailsWhenNotLocked(t *testing.T) {
	cfg := setupLockVault(t)

	if err := vault.Unlock(cfg); err == nil {
		t.Fatal("expected error when unlocking non-locked vault, got nil")
	}
}

func TestLock_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	if err := vault.Lock(cfg); err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestLockStatus_Unlocked(t *testing.T) {
	cfg := setupLockVault(t)
	status := vault.LockStatus(cfg)
	if status != "unlocked" {
		t.Fatalf("expected 'unlocked', got %q", status)
	}
}

func TestLockStatus_Locked(t *testing.T) {
	cfg := setupLockVault(t)
	_ = vault.Lock(cfg)
	status := vault.LockStatus(cfg)
	if len(status) == 0 || status == "unlocked" {
		t.Fatalf("expected non-empty locked status, got %q", status)
	}
	// lock file should still be present on disk
	if _, err := os.Stat(filepath.Join(cfg.VaultDir, ".envault.lock")); err != nil {
		t.Fatalf("lock file missing: %v", err)
	}
}
