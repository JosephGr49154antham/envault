package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRenameVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	src := filepath.Join(dir, ".env")
	return cfg, src
}

func writeRenameEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestRename_Success(t *testing.T) fg, src := setupRenameVault(t)
	writeRenameEnv(t, src, "DB_HOST=localhost\nDB_PORT=5432\nAPP_ENVproduction\n")

	if err := Rename(cfg, src, "DB_HOST", "DATABASE_HOST"); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	data := os.ReadFile(src)
	contents := string(data)
	if strings.Contains(contents, "DB_HOST=") {
		t.Error("old key still present after rename")
	}
	if !strings.Contains(contents, "DATABASE_HOST=localhost") {
		t.Error("new key not found after rename")
	}
}

func TestRename_KeyNotFound(t *testing.T) {
	cfg, src := setupRenameVault(t)
	writeRenameEnv(t, src, "APP_ENV=staging\n")

	err := Rename(cfg, src, "MISSING_KEY", "NEW_KEY")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRename_DuplicateNewKey(t *testing.T) {
	cfg, src := setupRenameVault(t)
	writeRenameEnv(t, src, "DB_HOST=localhost\nDATABASE_HOST=remote\n")

	err := Rename(cfg, src, "DB_HOST", "DATABASE_HOST")
	if err == nil {
		t.Fatal("expected error for duplicate key, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRename_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Rename(cfg, filepath.Join(dir, ".env"), "OLD", "NEW")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
	if !strings.Contains(err.Error(), "not initialised") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRename_SameKeyError(t *testing.T) {
	cfg, src := setupRenameVault(t)
	writeRenameEnv(t, src, "KEY=value\n")

	err := Rename(cfg, src, "KEY", "KEY")
	if err == nil {
		t.Fatal("expected error when old and new key are identical")
	}
}
