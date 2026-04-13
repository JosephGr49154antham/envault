package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/vault"
)

func setupCloneVault(t *testing.T) vault.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", ".env.age"),
	}
	); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg
}

func writeCloneEnv(t *testing.T, path,.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

opiesFile(t *testing.T) {
	cfg := setupCloneVault(t)
	src := filepath.Join(cfg.VaultDir, ".env.staging")
	writeCloneEnv(t, src, "DB_HOST=staging-db\nAPI_KEY=abc123\n")

	if err := vault.Clone(cfg, ".env.staging", ".env.production", false); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	dst := filepath.Join(cfg.VaultDir, ".env.production")
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "DB_HOST=staging-db\nAPI_KEY=abc123\n" {
		t.Errorf("unexpected content: %s", string(data))
	}
}

func TestClone_NoOverwrite(t *testing.T) {
	cfg := setupCloneVault(t)
	src := filepath.Join(cfg.VaultDir, ".env.staging")
	writeCloneEnv(t, src, "KEY=val\n")
	dst := filepath.Join(cfg.VaultDir, ".env.production")
	writeCloneEnv(t, dst, "KEY=other\n")

	err := vault.Clone(cfg, ".env.staging", ".env.production", false)
	if err == nil {
		t.Fatal("expected error when destination exists without overwrite")
	}
}

func TestClone_Overwrite(t *testing.T) {
	cfg := setupCloneVault(t)
	src := filepath.Join(cfg.VaultDir, ".env.staging")
	writeCloneEnv(t, src, "KEY=new\n")
	dst := filepath.Join(cfg.VaultDir, ".env.production")
	writeCloneEnv(t, dst, "KEY=old\n")

	if err := vault.Clone(cfg, ".env.staging", ".env.production", true); err != nil {
		t.Fatalf("Clone with overwrite: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "KEY=new\n" {
		t.Errorf("expected overwritten content, got: %s", string(data))
	}
}

func TestClone_SourceNotFound(t *testing.T) {
	cfg := setupCloneVault(t)
	err := vault.Clone(cfg, ".env.missing", ".env.production", false)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestClone_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", ".env.age"),
	}
	err := vault.Clone(cfg, ".env.staging", ".env.production", false)
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
