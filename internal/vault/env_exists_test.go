package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupExistsVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      dir,
		RecipientsFile: filepath.Join(dir, "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeExistsEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestExists_KeyFound(t *testing.T) {
	cfg, dir := setupExistsVault(t)
	src := filepath.Join(dir, ".env")
	writeExistsEnv(t, src, "APP_NAME=envault\nDEBUG=true\n")

	results, err := Exists(cfg, src, []string{"APP_NAME"}, ExistsOptions{ShowValue: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Found {
		t.Errorf("expected APP_NAME to be found")
	}
	if results[0].Value != "envault" {
		t.Errorf("expected value 'envault', got %q", results[0].Value)
	}
}

func TestExists_KeyNotFound(t *testing.T) {
	cfg, dir := setupExistsVault(t)
	src := filepath.Join(dir, ".env")
	writeExistsEnv(t, src, "APP_NAME=envault\n")

	results, err := Exists(cfg, src, []string{"MISSING_KEY"}, ExistsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Found {
		t.Errorf("expected MISSING_KEY to not be found")
	}
}

func TestExists_HidesValueWhenShowValueFalse(t *testing.T) {
	cfg, dir := setupExistsVault(t)
	src := filepath.Join(dir, ".env")
	writeExistsEnv(t, src, "SECRET=hunter2\n")

	results, err := Exists(cfg, src, []string{"SECRET"}, ExistsOptions{ShowValue: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results[0].Found {
		t.Errorf("expected SECRET to be found")
	}
	if results[0].Value != "" {
		t.Errorf("expected value to be hidden, got %q", results[0].Value)
	}
}

func TestExists_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir()}
	_, err := Exists(cfg, ".env", []string{"KEY"}, ExistsOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestExists_FileNotFound(t *testing.T) {
	cfg, dir := setupExistsVault(t)
	_, err := Exists(cfg, filepath.Join(dir, "missing.env"), []string{"KEY"}, ExistsOptions{})
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}
