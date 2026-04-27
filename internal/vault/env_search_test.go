package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupSearchVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeSearchEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestSearch_MatchesKey(t *testing.T) {
	cfg, dir := setupSearchVault(t)
	envPath := filepath.Join(dir, ".env")
	writeSearchEnv(t, envPath, "DATABASE_URL=postgres://localhost\nSECRET_KEY=abc123\nPORT=8080\n")

	results, err := Search(cfg, SearchOptions{Pattern: "DATABASE", CaseSensitive: true}, envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "DATABASE_URL" {
		t.Errorf("expected key DATABASE_URL, got %s", results[0].Key)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	cfg, dir := setupSearchVault(t)
	envPath := filepath.Join(dir, ".env")
	writeSearchEnv(t, envPath, "SECRET_KEY=abc\nPORT=9000\n")

	results, err := Search(cfg, SearchOptions{Pattern: "secret", CaseSensitive: false}, envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Key != "SECRET_KEY" {
		t.Errorf("expected SECRET_KEY match, got %+v", results)
	}
}

func TestSearch_SearchValues(t *testing.T) {
	cfg, dir := setupSearchVault(t)
	envPath := filepath.Join(dir, ".env")
	writeSearchEnv(t, envPath, "DB=postgres://localhost\nCACHE=redis://localhost\n")

	results, err := Search(cfg, SearchOptions{Pattern: "localhost", SearchValues: true}, envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/missing"}
	_, err := Search(cfg, SearchOptions{Pattern: "KEY"})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestSearch_EmptyPattern(t *testing.T) {
	cfg, dir := setupSearchVault(t)
	envPath := filepath.Join(dir, ".env")
	writeSearchEnv(t, envPath, "KEY=value\n")

	_, err := Search(cfg, SearchOptions{Pattern: ""}, envPath)
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestSearch_SkipsCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupSearchVault(t)
	envPath := filepath.Join(dir, ".env")
	writeSearchEnv(t, envPath, "# comment\n\nKEY=value\n")

	results, err := Search(cfg, SearchOptions{Pattern: "comment"}, envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for comment line, got %d", len(results))
	}
}
