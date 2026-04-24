package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupValuesVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeValuesEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestValues_AllValues(t *testing.T) {
	cfg, _ := setupValuesVault(t)
	writeValuesEnv(t, cfg.PlainFile, "# header\nDB_HOST=localhost\nDB_PORT=5432\nAPP_ENV=production\n")

	res, err := Values(cfg, ValuesOptions{})
	if err != nil {
		t.Fatalf("Values: %v", err)
	}
	if len(res.Values) != 3 {
		t.Fatalf("expected 3 values, got %d: %v", len(res.Values), res.Values)
	}
	if res.Values[0] != "localhost" || res.Values[1] != "5432" || res.Values[2] != "production" {
		t.Errorf("unexpected values: %v", res.Values)
	}
}

func TestValues_SelectedKeys(t *testing.T) {
	cfg, _ := setupValuesVault(t)
	writeValuesEnv(t, cfg.PlainFile, "DB_HOST=localhost\nDB_PORT=5432\nAPP_ENV=production\n")

	res, err := Values(cfg, ValuesOptions{Keys: []string{"DB_HOST", "APP_ENV"}})
	if err != nil {
		t.Fatalf("Values: %v", err)
	}
	if len(res.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(res.Values))
	}
	if res.Values[0] != "localhost" || res.Values[1] != "production" {
		t.Errorf("unexpected values: %v", res.Values)
	}
}

func TestValues_Quoted(t *testing.T) {
	cfg, _ := setupValuesVault(t)
	writeValuesEnv(t, cfg.PlainFile, "GREETING=hello world\n")

	res, err := Values(cfg, ValuesOptions{Quoted: true})
	if err != nil {
		t.Fatalf("Values: %v", err)
	}
	if len(res.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(res.Values))
	}
	if res.Values[0] != `"hello world"` {
		t.Errorf("expected quoted value, got %s", res.Values[0])
	}
}

func TestValues_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	_, err := Values(cfg, ValuesOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestValues_SkipsCommentsAndBlanks(t *testing.T) {
	cfg, _ := setupValuesVault(t)
	writeValuesEnv(t, cfg.PlainFile, "# comment\n\nKEY=value\n")

	res, err := Values(cfg, ValuesOptions{})
	if err != nil {
		t.Fatalf("Values: %v", err)
	}
	if len(res.Values) != 1 || res.Values[0] != "value" {
		t.Errorf("unexpected values: %v", res.Values)
	}
}
