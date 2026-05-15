package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRenameKeyVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeRenameKeyEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
}

func TestRenameKey_Success(t *testing.T) {
	cfg, dir := setupRenameKeyVault(t)
	src := filepath.Join(dir, ".env")
	writeRenameKeyEnv(t, src, "OLD_KEY=hello\nOTHER=world\n")

	err := RenameKey(cfg, RenameKeyOptions{Src: src, OldKey: "OLD_KEY", NewKey: "NEW_KEY"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(src)
	if !strings.Contains(string(data), "NEW_KEY=hello") {
		t.Errorf("expected NEW_KEY=hello in output, got:\n%s", data)
	}
	if strings.Contains(string(data), "OLD_KEY") {
		t.Errorf("OLD_KEY should have been removed, got:\n%s", data)
	}
}

func TestRenameKey_KeyNotFound(t *testing.T) {
	cfg, dir := setupRenameKeyVault(t)
	src := filepath.Join(dir, ".env")
	writeRenameKeyEnv(t, src, "SOME_KEY=value\n")

	err := RenameKey(cfg, RenameKeyOptions{Src: src, OldKey: "MISSING", NewKey: "NEW_KEY"})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRenameKey_NoOverwriteDst(t *testing.T) {
	cfg, dir := setupRenameKeyVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.new")
	writeRenameKeyEnv(t, src, "KEY=val\n")
	writeRenameKeyEnv(t, dst, "existing=content\n")

	err := RenameKey(cfg, RenameKeyOptions{Src: src, Dst: dst, OldKey: "KEY", NewKey: "KEY2"})
	if err == nil {
		t.Fatal("expected error when dst exists without --force")
	}
}

func TestRenameKey_ForceOverwriteDst(t *testing.T) {
	cfg, dir := setupRenameKeyVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.new")
	writeRenameKeyEnv(t, src, "KEY=val\n")
	writeRenameKeyEnv(t, dst, "existing=content\n")

	err := RenameKey(cfg, RenameKeyOptions{Src: src, Dst: dst, OldKey: "KEY", NewKey: "KEY2", Force: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "KEY2=val") {
		t.Errorf("expected KEY2=val in dst, got:\n%s", data)
	}
}

func TestRenameKey_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/missing"}
	err := RenameKey(cfg, RenameKeyOptions{Src: "any", OldKey: "A", NewKey: "B"})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
