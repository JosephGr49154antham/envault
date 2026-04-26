package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCastVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncFile:     filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		IdentityFile:   filepath.Join(dir, ".envault", "identity.age"),
	}
	if err := os.MkdirAll(cfg.VaultDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return cfg, dir
}

func writeCastEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCast_NormalisesBooleans(t *testing.T) {
	cfg, _ := setupCastVault(t)
	writeCastEnv(t, cfg.PlainFile, "DEBUG=yes\nVERBOSE=NO\nENABLED=1\n")

	results, err := Cast(cfg, CastOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Type != "bool" {
			t.Errorf("key %s: expected bool, got %s", r.Key, r.Type)
		}
		if r.Cast != "true" && r.Cast != "false" {
			t.Errorf("key %s: unexpected cast value %q", r.Key, r.Cast)
		}
	}
}

func TestCast_NormalisesIntegers(t *testing.T) {
	cfg, _ := setupCastVault(t)
	writeCastEnv(t, cfg.PlainFile, "PORT=08080\nTIMEOUT=30\n")

	results, err := Cast(cfg, CastOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Type != "int" {
			t.Errorf("key %s: expected int, got %s", r.Key, r.Type)
		}
	}
	// Leading zero should be stripped
	if results[0].Cast == "08080" {
		t.Errorf("leading zero not stripped: %s", results[0].Cast)
	}
}

func TestCast_SelectedKeys(t *testing.T) {
	cfg, _ := setupCastVault(t)
	writeCastEnv(t, cfg.PlainFile, "FLAG=yes\nNAME=alice\n")

	results, err := Cast(cfg, CastOptions{Keys: []string{"FLAG"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Key != "FLAG" {
		t.Fatalf("expected only FLAG in results, got %v", results)
	}
	// NAME should be unchanged in output file
	data, _ := os.ReadFile(cfg.PlainFile)
	if !strings.Contains(string(data), "NAME=alice") {
		t.Errorf("NAME should be unchanged: %s", string(data))
	}
}

func TestCast_CustomDst(t *testing.T) {
	cfg, dir := setupCastVault(t)
	writeCastEnv(t, cfg.PlainFile, "ACTIVE=true\n")
	dst := filepath.Join(dir, ".env.cast")

	_, err := Cast(cfg, CastOptions{Dst: dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected dst file to exist: %v", err)
	}
}

func TestCast_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	_, err := Cast(cfg, CastOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
