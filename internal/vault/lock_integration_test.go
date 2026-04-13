package vault_test

import (
	"path/filepath"
	"testing"

	"github.com/roamingthings/envault/internal/vault"
)

// TestLockUnlock_FullCycle exercises the complete lock → unlock → re-lock cycle.
func TestLockUnlock_FullCycle(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Initially unlocked
	if vault.IsLocked(cfg) {
		t.Fatal("new vault should not be locked")
	}

	// Lock
	if err := vault.Lock(cfg); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if !vault.IsLocked(cfg) {
		t.Fatal("vault should be locked after Lock()")
	}

	// Double-lock must fail
	if err := vault.Lock(cfg); err == nil {
		t.Fatal("double Lock should return error")
	}

	// Unlock
	if err := vault.Unlock(cfg); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if vault.IsLocked(cfg) {
		t.Fatal("vault should be unlocked after Unlock()")
	}

	// Double-unlock must fail
	if err := vault.Unlock(cfg); err == nil {
		t.Fatal("double Unlock should return error")
	}

	// Can lock again after unlock
	if err := vault.Lock(cfg); err != nil {
		t.Fatalf("re-Lock after Unlock: %v", err)
	}
	if !vault.IsLocked(cfg) {
		t.Fatal("vault should be locked again")
	}
}
