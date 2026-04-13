package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/recipients"
	"github.com/nicholasgasior/envault/internal/vault"
)

func setupInjectVault(t *testing.T) (vault.Config, *keymgr.Identity) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		IdentityFile:  filepath.Join(dir, "identity.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	if err := keymgr.SaveIdentity(id, cfg.IdentityFile); err != nil {
		t.Fatalf("save identity: %v", err)
	}
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}
	return cfg, id
}

func writeInjectEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestInject_SetsEnvVars(t *testing.T) {
	cfg, id := setupInjectVault(t)
	plain := "FOO=bar\nBAZ=qux\n"
	rec := []string{id.Recipient().String()}
	if err := crypto.EncryptFile([]byte(plain), cfg.EncryptedFile, rec); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	t.Setenv("FOO", "")
	t.Setenv("BAZ", "")
	os.Unsetenv("FOO")
	os.Unsetenv("BAZ")
	res, err := vault.Inject(cfg, vault.InjectOptions{})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	if len(res.Set) != 2 {
		t.Fatalf("expected 2 set, got %d", len(res.Set))
	}
	if os.Getenv("FOO") != "bar" {
		t.Errorf("FOO = %q, want bar", os.Getenv("FOO"))
	}
	if os.Getenv("BAZ") != "qux" {
		t.Errorf("BAZ = %q, want qux", os.Getenv("BAZ"))
	}
}

func TestInject_SkipsExistingWithoutOverwrite(t *testing.T) {
	cfg, id := setupInjectVault(t)
	plain := "EXISTING=new\n"
	rec := []string{id.Recipient().String()}
	if err := crypto.EncryptFile([]byte(plain), cfg.EncryptedFile, rec); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	t.Setenv("EXISTING", "original")
	res, err := vault.Inject(cfg, vault.InjectOptions{Overwrite: false})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	if len(res.Skipped) != 1 {
		t.Fatalf("expected 1 skipped, got %d", len(res.Skipped))
	}
	if os.Getenv("EXISTING") != "original" {
		t.Errorf("expected original value preserved")
	}
}

func TestInject_NotInitialised(t *testing.T) {
	cfg := vault.Config{VaultDir: t.TempDir() + "/missing"}
	_, err := vault.Inject(cfg, vault.InjectOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestInject_DryRun(t *testing.T) {
	cfg, id := setupInjectVault(t)
	plain := "DRYKEY=dryval\n"
	rec := []string{id.Recipient().String()}
	if err := crypto.EncryptFile([]byte(plain), cfg.EncryptedFile, rec); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	os.Unsetenv("DRYKEY")
	res, err := vault.Inject(cfg, vault.InjectOptions{DryRun: true})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	if len(res.Set) != 1 {
		t.Fatalf("expected 1 in set list, got %d", len(res.Set))
	}
	if _, ok := os.LookupEnv("DRYKEY"); ok {
		t.Error("DRYKEY should not be set in dry-run mode")
	}
}
