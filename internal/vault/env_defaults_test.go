package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/vault"
)

func setupDefaultsVault(t *testing.T) (string, vault.Config) {
	t.Helper()
	dir := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir, cfg
}

func writeDefaultsEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestDefaults_FillsMissingKeys(t *testing.T) {
	dir, cfg := setupDefaultsVault(t)
	_ = cfg

	src := filepath.Join(dir, ".env")
	defaults := filepath.Join(dir, ".env.defaults")
	dst := filepath.Join(dir, ".env.out")

	writeDefaultsEnv(t, src, "APP_NAME=myapp\nDEBUG=true\n")
	writeDefaultsEnv(t, defaults, "APP_NAME=default-app\nLOG_LEVEL=info\nTIMEOUT=30\n")

	if err := vault.Defaults(src, defaults, dst, false); err != nil {
		t.Fatalf("Defaults: %v", err)
	}

	out, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}

	content := string(out)
	// APP_NAME already set — should keep original value
	if !contains(content, "APP_NAME=myapp") {
		t.Errorf("expected APP_NAME=myapp in output, got:\n%s", content)
	}
	// LOG_LEVEL missing from src — should be filled from defaults
	if !contains(content, "LOG_LEVEL=info") {
		t.Errorf("expected LOG_LEVEL=info in output, got:\n%s", content)
	}
	// TIMEOUT missing from src — should be filled from defaults
	if !contains(content, "TIMEOUT=30") {
		t.Errorf("expected TIMEOUT=30 in output, got:\n%s", content)
	}
}

func TestDefaults_DefaultDst(t *testing.T) {
	dir, cfg := setupDefaultsVault(t)
	_ = cfg

	src := filepath.Join(dir, ".env")
	defaults := filepath.Join(dir, ".env.defaults")

	writeDefaultsEnv(t, src, "KEY=value\n")
	writeDefaultsEnv(t, defaults, "KEY=default\nEXTRA=extra\n")

	// dst = "" means write back to src
	if err := vault.Defaults(src, defaults, "", false); err != nil {
		t.Fatalf("Defaults: %v", err)
	}

	out, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read src: %v", err)
	}
	content := string(out)
	if !contains(content, "EXTRA=extra") {
		t.Errorf("expected EXTRA=extra written back to src, got:\n%s", content)
	}
}

func TestDefaults_NoOverwrite(t *testing.T) {
	dir, cfg := setupDefaultsVault(t)
	_ = cfg

	src := filepath.Join(dir, ".env")
	defaults := filepath.Join(dir, ".env.defaults")
	dst := filepath.Join(dir, ".env.out")

	writeDefaultsEnv(t, src, "KEY=original\n")
	writeDefaultsEnv(t, defaults, "KEY=default\n")
	writeDefaultsEnv(t, dst, "existing=content\n")

	err := vault.Defaults(src, defaults, dst, false)
	if err == nil {
		t.Fatal("expected error when dst exists and overwrite=false")
	}
}

func TestDefaults_OverwriteFlag(t *testing.T) {
	dir, cfg := setupDefaultsVault(t)
	_ = cfg

	src := filepath.Join(dir, ".env")
	defaults := filepath.Join(dir, ".env.defaults")
	dst := filepath.Join(dir, ".env.out")

	writeDefaultsEnv(t, src, "KEY=original\n")
	writeDefaultsEnv(t, defaults, "KEY=default\nNEW=added\n")
	writeDefaultsEnv(t, dst, "old=content\n")

	if err := vault.Defaults(src, defaults, dst, true); err != nil {
		t.Fatalf("Defaults with overwrite: %v", err)
	}

	out, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !contains(string(out), "NEW=added") {
		t.Errorf("expected NEW=added after overwrite, got:\n%s", string(out))
	}
}

func TestDefaults_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, ".env")
	defaults := filepath.Join(dir, ".env.defaults")

	_ = os.WriteFile(src, []byte("K=V\n"), 0o600)
	_ = os.WriteFile(defaults, []byte("K=default\n"), 0o600)

	// vault not initialised — Defaults should still work as a file operation
	// (Defaults does not require vault init)
	if err := vault.Defaults(src, defaults, "", false); err != nil {
		t.Fatalf("unexpected error for uninitialised vault: %v", err)
	}
}

// contains is a helper shared across test files in this package.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
