package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupLintVault(t *testing.T) (Config, string) {
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
	return cfg, dir
}

func writeEnvFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("writeEnvFile: %v", err)
	}
	return p
}

func TestLint_CleanFile(t *testing.T) {
	cfg, dir := setupLintVault(t)
	src := writeEnvFile(t, dir, ".env", "# comment\nDB_HOST=localhost\nDB_PORT=5432\n")

	issues, err := Lint(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestLint_DetectsBadKeyCase(t *testing.T) {
	cfg, dir := setupLintVault(t)
	src := writeEnvFile(t, dir, ".env", "db_host=localhost\n")

	issues, err := Lint(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Line != 1 {
		t.Errorf("expected line 1, got %d", issues[0].Line)
	}
}

func TestLint_DetectsDuplicateKey(t *testing.T) {
	cfg, dir := setupLintVault(t)
	src := writeEnvFile(t, dir, ".env", "FOO=bar\nFOO=baz\n")

	issues, err := Lint(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Line != 2 {
		t.Errorf("expected line 2, got %d", issues[0].Line)
	}
}

func TestLint_DetectsInvalidFormat(t *testing.T) {
	cfg, dir := setupLintVault(t)
	src := writeEnvFile(t, dir, ".env", "NOTAKVPAIR\n")

	issues, err := Lint(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestLint_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Lint(cfg, filepath.Join(dir, ".env"))
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
