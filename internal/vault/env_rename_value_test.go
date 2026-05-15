package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRenameValueVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := os.MkdirAll(cfg.VaultDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return cfg, dir
}

func writeRenameValueEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRenameValue_ReplacesMatchingValue(t *testing.T) {
	cfg, dir := setupRenameValueVault(t)
	src := filepath.Join(dir, ".env")
	writeRenameValueEnv(t, src, "DB_PASS=secret\nAPI_KEY=secret\nHOST=localhost\n")

	err := RenameValue(cfg, RenameValueOptions{Src: src, OldValue: "secret", NewValue: "hunter2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(src)
	got := string(data)
	if !strings.Contains(got, "DB_PASS=hunter2") {
		t.Errorf("expected DB_PASS=hunter2, got:\n%s", got)
	}
	if !strings.Contains(got, "API_KEY=hunter2") {
		t.Errorf("expected API_KEY=hunter2, got:\n%s", got)
	}
	if !strings.Contains(got, "HOST=localhost") {
		t.Errorf("HOST should be unchanged, got:\n%s", got)
	}
}

func TestRenameValue_CustomDst(t *testing.T) {
	cfg, dir := setupRenameValueVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeRenameValueEnv(t, src, "TOKEN=old\n")

	err := RenameValue(cfg, RenameValueOptions{Src: src, Dst: dst, OldValue: "old", NewValue: "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "TOKEN=new") {
		t.Errorf("expected TOKEN=new in dst, got: %s", data)
	}
}

func TestRenameValue_NoOverwrite(t *testing.T) {
	cfg, dir := setupRenameValueVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeRenameValueEnv(t, src, "K=v\n")
	writeRenameValueEnv(t, dst, "existing\n")

	err := RenameValue(cfg, RenameValueOptions{Src: src, Dst: dst, OldValue: "v", NewValue: "x"})
	if err == nil {
		t.Fatal("expected error when dst exists without --overwrite")
	}
}

func TestRenameValue_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := RenameValue(cfg, RenameValueOptions{Src: ".env", OldValue: "a", NewValue: "b"})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
