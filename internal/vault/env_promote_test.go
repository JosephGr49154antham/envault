package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupPromoteVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:        filepath.Join(dir, ".envault"),
		RecipientsFile:  filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:   filepath.Join(dir, ".envault", "env.age"),
		IdentityFile:    filepath.Join(dir, ".envault", "identity.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writePromoteEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestPromote_AddsNewKeys(t *testing.T) {
	cfg, dir := setupPromoteVault(t)
	src := filepath.Join(dir, ".env.staging")
	dst := filepath.Join(dir, ".env.production")

	writePromoteEnv(t, src, "DB_HOST=staging.db\nDB_PORT=5432\nNEW_KEY=hello\n")
	writePromoteEnv(t, dst, "DB_HOST=prod.db\nDB_PORT=5432\n")

	res, err := Promote(cfg, src, dst)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}

	if len(res.KeysAdded) != 1 || res.KeysAdded[0] != "NEW_KEY" {
		t.Errorf("expected NEW_KEY added, got %v", res.KeysAdded)
	}
	if len(res.KeysKept) != 2 {
		t.Errorf("expected 2 kept keys, got %v", res.KeysKept)
	}

	dstMap, _ := readEnvFileToMap(dst)
	if dstMap["DB_HOST"] != "prod.db" {
		t.Errorf("existing key overwritten: got %s", dstMap["DB_HOST"])
	}
	if dstMap["NEW_KEY"] != "hello" {
		t.Errorf("new key not written: got %s", dstMap["NEW_KEY"])
	}
}

func TestPromote_CreatesDstIfMissing(t *testing.T) {
	cfg, dir := setupPromoteVault(t)
	src := filepath.Join(dir, ".env.staging")
	dst := filepath.Join(dir, ".env.production")

	writePromoteEnv(t, src, "API_KEY=secret\n")

	res, err := Promote(cfg, src, dst)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if len(res.KeysAdded) != 1 {
		t.Errorf("expected 1 added key, got %v", res.KeysAdded)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("dst not created: %v", err)
	}
}

func TestPromote_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig(dir)
	_, err := Promote(cfg, filepath.Join(dir, "a"), filepath.Join(dir, "b"))
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestPromote_NoKeysAdded(t *testing.T) {
	cfg, dir := setupPromoteVault(t)
	src := filepath.Join(dir, ".env.staging")
	dst := filepath.Join(dir, ".env.production")

	writePromoteEnv(t, src, "DB_HOST=staging.db\n")
	writePromoteEnv(t, dst, "DB_HOST=prod.db\n")

	res, err := Promote(cfg, src, dst)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if len(res.KeysAdded) != 0 {
		t.Errorf("expected no keys added, got %v", res.KeysAdded)
	}
	if len(res.KeysKept) != 1 {
		t.Errorf("expected 1 kept key, got %v", res.KeysKept)
	}
}
