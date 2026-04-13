package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/recipients"
)

func setupTemplateVault(t *testing.T) (Config, *age.X25519Identity) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		IdentityFile:  filepath.Join(dir, ".envault", "identity.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}

	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}

	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	plain := "DB_HOST=localhost\nDB_PORT=5432\nSECRET_KEY=supersecret\n"
	if err := os.WriteFile(cfg.PlainFile, []byte(plain), 0o600); err != nil {
		t.Fatalf("write plain: %v", err)
	}

	if err := crypto.EncryptEnvFile(cfg.PlainFile, cfg.EncryptedFile, []age.Recipient{id.Recipient()}); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if err := os.WriteFile(cfg.IdentityFile, []byte(id.String()+"\n"), 0o600); err != nil {
		t.Fatalf("write identity: %v", err)
	}

	return cfg, id
}

func TestGenerateTemplate_CreatesFile(t *testing.T) {
	cfg, _ := setupTemplateVault(t)

	dst, err := GenerateTemplate(cfg, "")
	if err != nil {
		t.Fatalf("GenerateTemplate: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	content := string(data)
	for _, key := range []string{"DB_HOST", "DB_PORT", "SECRET_KEY"} {
		if !strings.Contains(content, key+"=") {
			t.Errorf("expected key %q in template, got:\n%s", key, content)
		}
	}

	if strings.Contains(content, "localhost") || strings.Contains(content, "supersecret") {
		t.Error("template must not contain real values")
	}
}

func TestGenerateTemplate_CustomDst(t *testing.T) {
	cfg, _ := setupTemplateVault(t)
	customDst := filepath.Join(t.TempDir(), "custom.template")

	out, err := GenerateTemplate(cfg, customDst)
	if err != nil {
		t.Fatalf("GenerateTemplate: %v", err)
	}
	if out != customDst {
		t.Errorf("expected dst %q, got %q", customDst, out)
	}
	if _, err := os.Stat(customDst); os.IsNotExist(err) {
		t.Error("custom dst file was not created")
	}
}

func TestGenerateTemplate_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/missing"}
	_, err := GenerateTemplate(cfg, "")
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestGenerateTemplate_DefaultPath(t *testing.T) {
	cfg, _ := setupTemplateVault(t)

	dst, err := GenerateTemplate(cfg, "")
	if err != nil {
		t.Fatalf("GenerateTemplate: %v", err)
	}

	expected := filepath.Join(filepath.Dir(cfg.PlainFile), ".env.template")
	if dst != expected {
		t.Errorf("expected default dst %q, got %q", expected, dst)
	}
}
