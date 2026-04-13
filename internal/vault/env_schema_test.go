package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupSchemaVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	schemaPath := filepath.Join(dir, ".env.schema")
	return cfg, schemaPath
}

func writeSchemaFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
}

func writeEnvForSchema(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestValidateSchema_AllPresent(t *testing.T) {
	cfg, schemaPath := setupSchemaVault(t)
	writeEnvForSchema(t, cfg.PlainFile, "DB_HOST=localhost\nDB_PORT=5432\nAPP_KEY=secret\n")
	writeSchemaFile(t, schemaPath, "# required keys\n!DB_HOST\n!DB_PORT\nAPP_KEY\n")

	res, err := ValidateSchema(cfg, schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Missing) != 0 {
		t.Errorf("expected no missing keys, got %v", res.Missing)
	}
	if len(res.Extra) != 0 {
		t.Errorf("expected no extra keys, got %v", res.Extra)
	}
}

func TestValidateSchema_MissingRequired(t *testing.T) {
	cfg, schemaPath := setupSchemaVault(t)
	writeEnvForSchema(t, cfg.PlainFile, "DB_HOST=localhost\n")
	writeSchemaFile(t, schemaPath, "!DB_HOST\n!DB_PORT\n!APP_KEY\n")

	res, err := ValidateSchema(cfg, schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Missing) != 2 {
		t.Errorf("expected 2 missing, got %v", res.Missing)
	}
}

func TestValidateSchema_ExtraKeys(t *testing.T) {
	cfg, schemaPath := setupSchemaVault(t)
	writeEnvForSchema(t, cfg.PlainFile, "DB_HOST=localhost\nUNKNOWN_KEY=foo\n")
	writeSchemaFile(t, schemaPath, "!DB_HOST\n")

	res, err := ValidateSchema(cfg, schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Extra) != 1 || res.Extra[0] != "UNKNOWN_KEY" {
		t.Errorf("expected UNKNOWN_KEY as extra, got %v", res.Extra)
	}
}

func TestValidateSchema_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	_, err := ValidateSchema(cfg, filepath.Join(dir, ".env.schema"))
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestValidateSchema_SchemaNotFound(t *testing.T) {
	cfg, schemaPath := setupSchemaVault(t)
	writeEnvForSchema(t, cfg.PlainFile, "DB_HOST=localhost\n")
	// do not write schema file
	_, err := ValidateSchema(cfg, schemaPath)
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}
