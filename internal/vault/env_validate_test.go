package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupValidateVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeValidateEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestValidate_CleanFile(t *testing.T) {
	cfg, dir := setupValidateVault(t)
	src := filepath.Join(dir, ".env")
	writeValidateEnv(t, src, "APP_NAME=envault\nDEBUG=true\n# comment\n\nPORT=8080\n")

	res, err := Validate(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid, got issues: %v", res.Issues)
	}
	if len(res.Issues) != 0 {
		t.Errorf("expected no issues, got %d", len(res.Issues))
	}
}

func TestValidate_BadKeyCase(t *testing.T) {
	cfg, dir := setupValidateVault(t)
	src := filepath.Join(dir, ".env")
	writeValidateEnv(t, src, "appName=envault\nDEBUG=true\n")

	res, err := Validate(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Error("expected invalid due to bad key case")
	}
	if len(res.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(res.Issues))
	}
}

func TestValidate_EmptyValue(t *testing.T) {
	cfg, dir := setupValidateVault(t)
	src := filepath.Join(dir, ".env")
	writeValidateEnv(t, src, "TOKEN=\nDEBUG=true\n")

	res, err := Validate(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// empty value is a warning, not an error
	if !res.Valid {
		t.Error("expected valid (empty value is only a warning)")
	}
	if len(res.Issues) != 1 {
		t.Errorf("expected 1 warning issue, got %d", len(res.Issues))
	}
}

func TestValidate_InvalidLine(t *testing.T) {
	cfg, dir := setupValidateVault(t)
	src := filepath.Join(dir, ".env")
	writeValidateEnv(t, src, "VALID=yes\nthis is not valid\n")

	res, err := Validate(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Error("expected invalid due to malformed line")
	}
	if len(res.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(res.Issues))
	}
}

func TestValidate_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	_, err := Validate(cfg, "")
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestValidate_DefaultSrc(t *testing.T) {
	cfg, _ := setupValidateVault(t)
	writeValidateEnv(t, cfg.PlainFile, "KEY=value\n")

	res, err := Validate(cfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid, got issues: %v", res.Issues)
	}
}
