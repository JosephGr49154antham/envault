package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/envault/envault/internal/crypto"
	"github.com/envault/envault/internal/keymgr"
	"github.com/envault/envault/internal/recipients"
	"github.com/envault/envault/internal/vault"
)

func setupExportVault(t *testing.T) (vault.Config, *age.X25519Identity) {
	t.Helper()
	dir := t.TempDir()

	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", ".env.age"),
		PlainFile:     filepath.Join(dir, ".env"),
		IdentityFile:  filepath.Join(dir, ".envault", "identity.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	if err := keymgr.SaveIdentity(id, cfg.IdentityFile); err != nil {
		t.Fatalf("save identity: %v", err)
	}
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	plain := "DB_HOST=localhost\nDB_PORT=5432\n"
	if err := os.WriteFile(cfg.PlainFile, []byte(plain), 0o600); err != nil {
		t.Fatalf("write plain: %v", err)
	}
	if err := crypto.EncryptFile(cfg.PlainFile, cfg.EncryptedFile, []age.Recipient{id.Recipient()}); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return cfg, id
}

func TestExport_WritesDecryptedFile(t *testing.T) {
	cfg, _ := setupExportVault(t)
	out := filepath.Join(t.TempDir(), "exported.env")

	got, err := vault.Export(cfg, vault.ExportOptions{OutputPath: out})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if got != out {
		t.Errorf("returned path = %q, want %q", got, out)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "DB_HOST") {
		t.Errorf("exported file missing expected content, got: %s", data)
	}
}

func TestExport_DefaultOutputPath(t *testing.T) {
	cfg, _ := setupExportVault(t)

	got, err := vault.Export(cfg, vault.ExportOptions{})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if !strings.Contains(filepath.Base(got), "-export-") {
		t.Errorf("default output name %q missing timestamp segment", got)
	}
	if _, err := os.Stat(got); err != nil {
		t.Errorf("output file does not exist: %v", err)
	}
	// cleanup
	os.Remove(got)
}

func TestExport_NoOverwrite(t *testing.T) {
	cfg, _ := setupExportVault(t)
	out := filepath.Join(t.TempDir(), "exported.env")
	os.WriteFile(out, []byte("existing"), 0o600)

	_, err := vault.Export(cfg, vault.ExportOptions{OutputPath: out})
	if err == nil {
		t.Fatal("expected error when output exists and Overwrite=false")
	}
}

func TestExport_NotInitialised(t *testing.T) {
	cfg := vault.Config{
		VaultDir:      filepath.Join(t.TempDir(), ".envault"),
		EncryptedFile: "/nonexistent/.env.age",
		PlainFile:     "/nonexistent/.env",
	}
	_, err := vault.Export(cfg, vault.ExportOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
