package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupWhitelistVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeWhitelistEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestWhitelist_KeepsAllowedKeys(t *testing.T) {
	cfg, dir := setupWhitelistVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeWhitelistEnv(t, src, "DB_HOST=localhost\nDB_PORT=5432\nSECRET=hunter2\n")

	err := Whitelist(cfg, WhitelistOptions{Src: src, Dst: dst, Keys: []string{"DB_HOST", "DB_PORT"}, Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if strings.Contains(string(data), "SECRET") {
		t.Error("expected SECRET to be excluded")
	}
	if !strings.Contains(string(data), "DB_HOST") {
		t.Error("expected DB_HOST to be included")
	}
}

func TestWhitelist_DefaultDst(t *testing.T) {
	cfg, dir := setupWhitelistVault(t)
	src := filepath.Join(dir, ".env")
	writeWhitelistEnv(t, src, "FOO=bar\nBAZ=qux\n")

	err := Whitelist(cfg, WhitelistOptions{Src: src, Keys: []string{"FOO"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := src + ".whitelist"
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected default dst %q to exist", expected)
	}
}

func TestWhitelist_NoOverwrite(t *testing.T) {
	cfg, dir := setupWhitelistVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeWhitelistEnv(t, src, "A=1\n")
	os.WriteFile(dst, []byte("existing"), 0o600)

	err := Whitelist(cfg, WhitelistOptions{Src: src, Dst: dst, Keys: []string{"A"}, Overwrite: false})
	if err == nil {
		t.Error("expected error when dst exists and overwrite is false")
	}
}

func TestWhitelist_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{VaultDir: filepath.Join(dir, ".envault")}

	err := Whitelist(cfg, WhitelistOptions{Src: filepath.Join(dir, ".env"), Keys: []string{"X"}})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}
