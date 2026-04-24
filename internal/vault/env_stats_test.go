package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupStatsVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:        filepath.Join(dir, ".envault"),
		RecipientsFile:  filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:   filepath.Join(dir, ".envault", "env.age"),
		IdentityFile:    filepath.Join(dir, ".envault", "identity.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeStatsEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestStats_BasicCounts(t *testing.T) {
	cfg, dir := setupStatsVault(t)
	src := filepath.Join(dir, ".env")
	writeStatsEnv(t, src, "# header\nAPP_NAME=myapp\nDEBUG=true\nEMPTY=\n\nDB_HOST=localhost\n")

	res, err := Stats(cfg, src)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if res.TotalLines != 6 {
		t.Errorf("TotalLines: want 6, got %d", res.TotalLines)
	}
	if res.KeyValuePairs != 4 {
		t.Errorf("KeyValuePairs: want 4, got %d", res.KeyValuePairs)
	}
	if res.CommentLines != 1 {
		t.Errorf("CommentLines: want 1, got %d", res.CommentLines)
	}
	if res.BlankLines != 1 {
		t.Errorf("BlankLines: want 1, got %d", res.BlankLines)
	}
	if res.EmptyValues != 1 {
		t.Errorf("EmptyValues: want 1, got %d", res.EmptyValues)
	}
}

func TestStats_DuplicateKeys(t *testing.T) {
	cfg, dir := setupStatsVault(t)
	src := filepath.Join(dir, ".env")
	writeStatsEnv(t, src, "KEY=a\nKEY=b\nOTHER=c\n")

	res, err := Stats(cfg, src)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if res.DuplicateKeys != 1 {
		t.Errorf("DuplicateKeys: want 1, got %d", res.DuplicateKeys)
	}
	if res.UniqueKeys != 2 {
		t.Errorf("UniqueKeys: want 2, got %d", res.UniqueKeys)
	}
}

func TestStats_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
	}
	_, err := Stats(cfg, filepath.Join(dir, ".env"))
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestStats_EmptyFile(t *testing.T) {
	cfg, dir := setupStatsVault(t)
	src := filepath.Join(dir, ".env")
	writeStatsEnv(t, src, "")

	res, err := Stats(cfg, src)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if res.TotalLines != 0 {
		t.Errorf("TotalLines: want 0, got %d", res.TotalLines)
	}
	if res.KeyValuePairs != 0 {
		t.Errorf("KeyValuePairs: want 0, got %d", res.KeyValuePairs)
	}
}
