package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/vault"
)

func setupCountVault(t *testing.T) (vault.Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeCountEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	return p
}

func TestCount_BasicStats(t *testing.T) {
	cfg, dir := setupCountVault(t)
	src := writeCountEnv(t, dir, ".env", "# header\nDB_HOST=localhost\nDB_PASS=\nAPI_KEY=abc123\n\n# another comment\nPORT=8080\n")

	res, err := vault.Count(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 4 {
		t.Errorf("Total: got %d, want 4", res.Total)
	}
	if res.Set != 3 {
		t.Errorf("Set: got %d, want 3", res.Set)
	}
	if res.Empty != 1 {
		t.Errorf("Empty: got %d, want 1", res.Empty)
	}
	if res.Comments != 2 {
		t.Errorf("Comments: got %d, want 2", res.Comments)
	}
	if res.Blanks != 1 {
		t.Errorf("Blanks: got %d, want 1", res.Blanks)
	}
}

func TestCount_EmptyFile(t *testing.T) {
	cfg, dir := setupCountVault(t)
	src := writeCountEnv(t, dir, ".env", "")

	res, err := vault.Count(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 0 || res.Set != 0 || res.Empty != 0 {
		t.Errorf("expected all zeros for empty file, got %+v", res)
	}
}

func TestCount_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := vault.Count(cfg, ".env")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestCount_DefaultSrc(t *testing.T) {
	cfg, dir := setupCountVault(t)
	_ = writeCountEnv(t, dir, ".env", "KEY=value\n")

	// Change working directory so the default ".env" resolves correctly.
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)            //nolint:errcheck

	res, err := vault.Count(cfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total: got %d, want 1", res.Total)
	}
}
