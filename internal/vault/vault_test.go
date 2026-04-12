package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/vault"
)

func TestInit_CreatesVaultDir(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := vault.Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if _, err := os.Stat(cfg.VaultDir); err != nil {
		t.Errorf("vault dir not created: %v", err)
	}

	if _, err := os.Stat(cfg.RecipientsPath); err != nil {
		t.Errorf("recipients file not created: %v", err)
	}
}

func TestInit_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	if _, err := vault.Init(tmpDir); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	if _, err := vault.Init(tmpDir); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}
}

func TestIsInitialised(t *testing.T) {
	tmpDir := t.TempDir()

	if vault.IsInitialised(tmpDir) {
		t.Error("expected not initialised before Init")
	}

	if _, err := vault.Init(tmpDir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if !vault.IsInitialised(tmpDir) {
		t.Error("expected initialised after Init")
	}
}

func TestDefaultConfig_Paths(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := vault.DefaultConfig(tmpDir)

	expectedVaultDir := filepath.Join(tmpDir, vault.DefaultVaultDir)
	if cfg.VaultDir != expectedVaultDir {
		t.Errorf("V %q, want %q", cfg.VaultDir, expectedVaultDir)
	}

	if cfg.PlaintextPath != filepath.Join(tmpDir, ".env") {
		t.Errorf("PlaintextPath = %q", cfg.PlaintextPath)
	}

	if cfg.EncryptedPath != filepath.Join(expectedVaultDir, vault.EncryptedEnvFile) {
		t.Errorf("EncryptedPath = %q", cfg.EncryptedPath)
	}
}
