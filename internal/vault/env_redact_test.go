package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rodrwan/envault/internal/vault"
)

func setupRedactVault(t *testing.T) (vault.Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
		IdentityFile:  filepath.Join(dir, ".envault", "identity.txt"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeRedactEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return p
}

func TestRedact_SensitiveKeysAreRedacted(t *testing.T) {
	cfg, dir := setupRedactVault(t)
	src := writeRedactEnv(t, dir, ".env", "APP_NAME=myapp\nDB_PASSWORD=s3cr3t\nAPI_KEY=abc123\nPORT=8080\n")
	dst := filepath.Join(dir, ".env.redacted")

	result, err := vault.Redact(cfg, src, dst)
	if err != nil {
		t.Fatalf("Redact: %v", err)
	}

	if result.Total != 4 {
		t.Errorf("expected 4 total keys, got %d", result.Total)
	}
	if len(result.Redacted) != 2 {
		t.Errorf("expected 2 redacted keys, got %d: %v", len(result.Redacted), result.Redacted)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	body := string(data)
	if strings.Contains(body, "s3cr3t") {
		t.Error("expected DB_PASSWORD value to be redacted")
	}
	if strings.Contains(body, "abc123") {
		t.Error("expected API_KEY value to be redacted")
	}
	if !strings.Contains(body, "APP_NAME=myapp") {
		t.Error("expected APP_NAME to be preserved")
	}
	if !strings.Contains(body, "PORT=8080") {
		t.Error("expected PORT to be preserved")
	}
}

func TestRedact_DefaultDst(t *testing.T) {
	cfg, dir := setupRedactVault(t)
	src := writeRedactEnv(t, dir, ".env", "TOKEN=abc\n")

	_, err := vault.Redact(cfg, src, "")
	if err != nil {
		t.Fatalf("Redact: %v", err)
	}

	expected := src + ".redacted"
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected default dst %s to exist: %v", expected, err)
	}
}

func TestRedact_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
		IdentityFile:  filepath.Join(dir, ".envault", "identity.txt"),
	}
	_, err := vault.Redact(cfg, ".env", "")
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestRedact_PreservesCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupRedactVault(t)
	content := "# This is a comment\n\nAPP=hello\nSECRET_KEY=hunter2\n"
	src := writeRedactEnv(t, dir, ".env", content)
	dst := filepath.Join(dir, ".env.redacted")

	_, err := vault.Redact(cfg, src, dst)
	if err != nil {
		t.Fatalf("Redact: %v", err)
	}

	data, _ := os.ReadFile(dst)
	body := string(data)
	if !strings.Contains(body, "# This is a comment") {
		t.Error("expected comment to be preserved")
	}
	if !strings.Contains(body, "APP=hello") {
		t.Error("expected APP=hello to be preserved")
	}
	if strings.Contains(body, "hunter2") {
		t.Error("expected SECRET_KEY value to be redacted")
	}
}
