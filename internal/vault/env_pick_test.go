package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupPickVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writePickEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestPick_ExtractsSelectedKeys(t *testing.T) {
	cfg, dir := setupPickVault(t)
	src := filepath.Join(dir, ".env")
	writePickEnv(t, src, "APP_NAME=envault\nDB_HOST=localhost\nDB_PORT=5432\nSECRET=abc\n")

	dst := filepath.Join(dir, "partial.env")
	got, err := Pick(cfg, PickOptions{Keys: []string{"DB_HOST", "SECRET"}, Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if got != dst {
		t.Errorf("returned path = %q, want %q", got, dst)
	}

	data, _ := os.ReadFile(dst)
	body := string(data)
	if !strings.Contains(body, "DB_HOST=localhost") {
		t.Error("expected DB_HOST in output")
	}
	if !strings.Contains(body, "SECRET=abc") {
		t.Error("expected SECRET in output")
	}
	if strings.Contains(body, "APP_NAME") {
		t.Error("APP_NAME should not be in output")
	}
	if strings.Contains(body, "DB_PORT") {
		t.Error("DB_PORT should not be in output")
	}
}

func TestPick_DefaultDst(t *testing.T) {
	cfg, dir := setupPickVault(t)
	src := filepath.Join(dir, ".env")
	writePickEnv(t, src, "FOO=bar\nBAZ=qux\n")

	got, err := Pick(cfg, PickOptions{Keys: []string{"FOO"}, Src: src})
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	want := filepath.Join(dir, ".picked.env")
	if got != want {
		t.Errorf("default dst = %q, want %q", got, want)
	}
}

func TestPick_NoOverwrite(t *testing.T) {
	cfg, dir := setupPickVault(t)
	src := filepath.Join(dir, ".env")
	writePickEnv(t, src, "FOO=bar\n")
	dst := filepath.Join(dir, "out.env")
	writePickEnv(t, dst, "existing\n")

	_, err := Pick(cfg, PickOptions{Keys: []string{"FOO"}, Src: src, Dst: dst})
	if err == nil {
		t.Fatal("expected error when destination exists without --overwrite")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPick_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
	}
	_, err := Pick(cfg, PickOptions{Keys: []string{"FOO"}, Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
