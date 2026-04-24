package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupKeysVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeKeysEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestKeys_ReturnsKeys(t *testing.T) {
	cfg, dir := setupKeysVault(t)
	src := filepath.Join(dir, ".env")
	writeKeysEnv(t, src, "APP_NAME=envault\nDB_HOST=localhost\nSECRET=abc\n")

	keys, err := Keys(cfg, src, KeysOptions{})
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "APP_NAME" || keys[1] != "DB_HOST" || keys[2] != "SECRET" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestKeys_Sorted(t *testing.T) {
	cfg, dir := setupKeysVault(t)
	src := filepath.Join(dir, ".env")
	writeKeysEnv(t, src, "ZEBRA=1\nAPPLE=2\nMIDDLE=3\n")

	keys, err := Keys(cfg, src, KeysOptions{Sorted: true})
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if keys[0] != "APPLE" || keys[1] != "MIDDLE" || keys[2] != "ZEBRA" {
		t.Errorf("not sorted: %v", keys)
	}
}

func TestKeys_ValuesOnly(t *testing.T) {
	cfg, dir := setupKeysVault(t)
	src := filepath.Join(dir, ".env")
	writeKeysEnv(t, src, "HOST=localhost\nPORT=5432\n")

	vals, err := Keys(cfg, src, KeysOptions{ValuesOnly: true})
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(vals) != 2 || vals[0] != "localhost" || vals[1] != "5432" {
		t.Errorf("unexpected values: %v", vals)
	}
}

func TestKeys_SkipsCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupKeysVault(t)
	src := filepath.Join(dir, ".env")
	writeKeysEnv(t, src, "# comment\n\nKEY=val\n")

	keys, err := Keys(cfg, src, KeysOptions{})
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) != 1 || keys[0] != "KEY" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestKeys_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Keys(cfg, filepath.Join(dir, ".env"), KeysOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
