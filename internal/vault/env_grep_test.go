package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupGrepVault(t *testing.T) (Config, string) {
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

func writeGrepEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestGrep_MatchesKey(t *testing.T) {
	cfg, dir := setupGrepVault(t)
	envPath := filepath.Join(dir, ".env")
	writeGrepEnv(t, envPath, "DB_HOST=localhost\nDB_PORT=5432\nAPP_SECRET=abc\n")

	results, err := Grep(cfg, GrepOptions{Pattern: "^DB_"}, envPath)
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Key != "DB_HOST" || results[1].Key != "DB_PORT" {
		t.Errorf("unexpected keys: %v", results)
	}
}

func TestGrep_CaseInsensitive(t *testing.T) {
	cfg, dir := setupGrepVault(t)
	envPath := filepath.Join(dir, ".env")
	writeGrepEnv(t, envPath, "DB_HOST=localhost\ndb_user=admin\n")

	results, err := Grep(cfg, GrepOptions{Pattern: "db_", CaseSensitive: false}, envPath)
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 case-insensitive results, got %d", len(results))
	}
}

func TestGrep_SearchValues(t *testing.T) {
	cfg, dir := setupGrepVault(t)
	envPath := filepath.Join(dir, ".env")
	writeGrepEnv(t, envPath, "HOST=localhost\nURL=http://localhost:8080\nSECRET=xyz\n")

	results, err := Grep(cfg, GrepOptions{Pattern: "localhost", SearchValues: true}, envPath)
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 value matches, got %d", len(results))
	}
}

func TestGrep_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Grep(cfg, GrepOptions{Pattern: "DB"}, filepath.Join(dir, ".env"))
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestGrep_InvalidPattern(t *testing.T) {
	cfg, dir := setupGrepVault(t)
	envPath := filepath.Join(dir, ".env")
	writeGrepEnv(t, envPath, "KEY=val\n")

	_, err := Grep(cfg, GrepOptions{Pattern: "[invalid"}, envPath)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestGrep_SkipsCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupGrepVault(t)
	envPath := filepath.Join(dir, ".env")
	writeGrepEnv(t, envPath, "# DB_HOST comment\n\nDB_HOST=localhost\n")

	results, err := Grep(cfg, GrepOptions{Pattern: "DB_HOST"}, envPath)
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (comment skipped), got %d", len(results))
	}
	if results[0].LineNum != 3 {
		t.Errorf("expected line 3, got %d", results[0].LineNum)
	}
}
