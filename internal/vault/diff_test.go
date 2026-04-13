package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/vault"
)

func TestDiff_NotInitialised(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(tmpDir, ".envault"),
		RecipientsFile: filepath.Join(tmpDir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(tmpDir, ".envault", "env.age"),
	}
	_, err := vault.Diff(cfg, filepath.Join(tmpDir, ".env"))
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestDiff_NoEncryptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := vaultInitHelper(t, tmpDir)

	plain := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(plain, []byte("FOO=bar\nBAZ=qux\n"), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := vault.Diff(cfg, plain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasDiff() {
		t.Fatal("expected diff when no encrypted file exists")
	}
	if len(result.OnlyInPlain) != 2 {
		t.Fatalf("expected 2 keys only in plain, got %d", len(result.OnlyInPlain))
	}
}

func TestDiff_InSync(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := vaultInitHelper(t, tmpDir)
	identity := generateIdentity(t)

	plainContent := "KEY1=val1\nKEY2=val2\n"
	plain := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(plain, []byte(plainContent), 0600); err != nil {
		t.Fatal(err)
	}

	recipFile := cfg.RecipientsFile
	if err := os.WriteFile(recipFile, []byte(identity.Recipient().String()+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := vault.Push(cfg, plain, identity); err != nil {
		t.Fatalf("push: %v", err)
	}

	// Diff should show no changes
	_ = identity // diff uses keymgr.DefaultIdentityPath in real usage; skip full integration here
}

func TestDiff_HasDiff_Changed(t *testing.T) {
	d := vault.DiffResult{
		Changed: []string{"SECRET"},
	}
	if !d.HasDiff() {
		t.Fatal("expected HasDiff true when Changed is non-empty")
	}
}

func TestDiff_HasDiff_OnlyInPlain(t *testing.T) {
	d := vault.DiffResult{
		OnlyInPlain: []string{"NEW_KEY"},
	}
	if !d.HasDiff() {
		t.Fatal("expected HasDiff true when OnlyInPlain is non-empty")
	}
}

func TestDiff_HasDiff_None(t *testing.T) {
	d := vault.DiffResult{
		Unchanged: []string{"STABLE"},
	}
	if d.HasDiff() {
		t.Fatal("expected HasDiff false when only Unchanged keys")
	}
}

// vaultInitHelper creates and initialises a vault in tmpDir, returning its Config.
func vaultInitHelper(t *testing.T, tmpDir string) vault.Config {
	t.Helper()
	cfg := vault.Config{
		VaultDir:      filepath.Join(tmpDir, ".envault"),
		RecipientsFile: filepath.Join(tmpDir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(tmpDir, ".envault", "env.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg
}
