package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/nicholasgasior/envault/internal/vault"
)

func setupImportVault(t *testing.T) (vault.Config, *age.X25519Identity) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedFile:  filepath.Join(dir, ".envault", "secrets.env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	pub := id.Recipient().String()
	f, err := os.OpenFile(cfg.RecipientsFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open recipients: %v", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s\n", pub); err != nil {
		t.Fatalf("write recipient: %v", err)
	}
	return cfg, id
}

func TestImport_EncryptsSourceFile(t *testing.T) {
	cfg, _ := setupImportVault(t)

	src := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(src, []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := vault.Import(cfg, src, false); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if _, err := os.Stat(cfg.EncryptedFile); err != nil {
		t.Errorf("encrypted file not created: %v", err)
	}
}

func TestImport_NoOverwrite(t *testing.T) {
	cfg, _ := setupImportVault(t)

	src := filepath.Join(t.TempDir(), ".env")
	_ = os.WriteFile(src, []byte("A=1\n"), 0o600)

	// First import succeeds.
	if err := vault.Import(cfg, src, false); err != nil {
		t.Fatalf("first Import: %v", err)
	}

	// Second import without overwrite must fail.
	if err := vault.Import(cfg, src, false); err == nil {
		t.Error("expected error when overwrite=false and file exists")
	}
}

func TestImport_Overwrite(t *testing.T) {
	cfg, _ := setupImportVault(t)

	src := filepath.Join(t.TempDir(), ".env")
	_ = os.WriteFile(src, []byte("A=1\n"), 0o600)
	_ = vault.Import(cfg, src, false)

	if err := vault.Import(cfg, src, true); err != nil {
		t.Errorf("Import with overwrite: %v", err)
	}
}

func TestImport_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedFile:  filepath.Join(dir, ".envault", "secrets.env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	src := filepath.Join(dir, ".env")
	_ = os.WriteFile(src, []byte("A=1\n"), 0o600)

	if err := vault.Import(cfg, src, false); err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestImport_SourceNotFound(t *testing.T) {
	cfg, _ := setupImportVault(t)
	if err := vault.Import(cfg, "/nonexistent/.env", false); err == nil {
		t.Error("expected error for missing source file")
	}
}
