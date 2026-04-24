package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupContainsVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:        filepath.Join(dir, ".envault"),
		RecipientsFile:  filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:   filepath.Join(dir, ".envault", "env.age"),
		PlainFile:       filepath.Join(dir, ".env"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeContainsEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
}

func TestContains_KeyFound(t *testing.T) {
	cfg, dir := setupContainsVault(t)
	src := filepath.Join(dir, ".env")
	writeContainsEnv(t, src, "DB_HOST=localhost\nDB_PORT=5432\nSECRET=abc123\n")

	results, err := Contains(cfg, src, []string{"DB_HOST", "SECRET"}, ContainsOptions{ShowValue: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Found || results[0].Key != "DB_HOST" || results[0].Value != "localhost" {
		t.Errorf("unexpected result for DB_HOST: %+v", results[0])
	}
	if !results[1].Found || results[1].Key != "SECRET" || results[1].Value != "abc123" {
		t.Errorf("unexpected result for SECRET: %+v", results[1])
	}
}

func TestContains_KeyNotFound(t *testing.T) {
	cfg, dir := setupContainsVault(t)
	src := filepath.Join(dir, ".env")
	writeContainsEnv(t, src, "EXISTING=yes\n")

	results, err := Contains(cfg, src, []string{"MISSING"}, ContainsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Found {
		t.Errorf("expected MISSING to not be found")
	}
}

func TestContains_HidesValueWhenShowValueFalse(t *testing.T) {
	cfg, dir := setupContainsVault(t)
	src := filepath.Join(dir, ".env")
	writeContainsEnv(t, src, "TOKEN=supersecret\n")

	results, err := Contains(cfg, src, []string{"TOKEN"}, ContainsOptions{ShowValue: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Value != "" {
		t.Errorf("expected value to be hidden, got %q", results[0].Value)
	}
	if !results[0].Found {
		t.Errorf("expected TOKEN to be found")
	}
}

func TestContains_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
		PlainFile:      filepath.Join(dir, ".env"),
	}
	_, err := Contains(cfg, filepath.Join(dir, ".env"), []string{"KEY"}, ContainsOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestContains_NoKeys(t *testing.T) {
	cfg, dir := setupContainsVault(t)
	src := filepath.Join(dir, ".env")
	writeContainsEnv(t, src, "A=1\n")

	_, err := Contains(cfg, src, []string{}, ContainsOptions{})
	if err == nil {
		t.Fatal("expected error when no keys provided")
	}
}
