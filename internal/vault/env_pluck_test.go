package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupPluckVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writePluckEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestPluck_ReturnsKeyValue(t *testing.T) {
	cfg, dir := setupPluckVault(t)
	src := filepath.Join(dir, ".env")
	writePluckEnv(t, src, "FOO=bar\nBAZ=qux\nSECRET=hunter2\n")

	results, err := Pluck(cfg, PluckOptions{Src: src, Keys: []string{"BAZ", "FOO"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "BAZ=qux" {
		t.Errorf("expected BAZ=qux, got %s", results[0])
	}
	if results[1] != "FOO=bar" {
		t.Errorf("expected FOO=bar, got %s", results[1])
	}
}

func TestPluck_RawMode(t *testing.T) {
	cfg, dir := setupPluckVault(t)
	src := filepath.Join(dir, ".env")
	writePluckEnv(t, src, "TOKEN=abc123\n")

	results, err := Pluck(cfg, PluckOptions{Src: src, Keys: []string{"TOKEN"}, Raw: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0] != "abc123" {
		t.Errorf("expected bare value 'abc123', got %v", results)
	}
}

func TestPluck_KeyNotFound(t *testing.T) {
	cfg, dir := setupPluckVault(t)
	src := filepath.Join(dir, ".env")
	writePluckEnv(t, src, "FOO=bar\n")

	_, err := Pluck(cfg, PluckOptions{Src: src, Keys: []string{"MISSING"}})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestPluck_DefaultSrc(t *testing.T) {
	cfg, _ := setupPluckVault(t)
	writePluckEnv(t, cfg.PlainFile, "DB_URL=postgres://localhost/dev\n")

	results, err := Pluck(cfg, PluckOptions{Keys: []string{"DB_URL"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0] != "DB_URL=postgres://localhost/dev" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestPluck_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	_, err := Pluck(cfg, PluckOptions{Keys: []string{"FOO"}})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
