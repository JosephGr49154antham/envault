package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupUnsetVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	src := filepath.Join(dir, ".env")
	return cfg, src
}

func writeUnsetEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestUnset_RemovesKey(t *testing.T) {
	cfg, src := setupUnsetVault(t)
	writeUnsetEnv(t, src, "FOO=bar\nBAZ=qux\nQUUX=hello\n")

	res, err := Unset(cfg, UnsetOptions{Src: src, Keys: []string{"BAZ"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Removed) != 1 || res.Removed[0] != "BAZ" {
		t.Errorf("expected BAZ removed, got %v", res.Removed)
	}
	data, _ := os.ReadFile(src)
	if strings.Contains(string(data), "BAZ") {
		t.Error("BAZ should have been removed from file")
	}
}

func TestUnset_ReportsMissingKey(t *testing.T) {
	cfg, src := setupUnsetVault(t)
	writeUnsetEnv(t, src, "FOO=bar\n")

	res, err := Unset(cfg, UnsetOptions{Src: src, Keys: []string{"NOPE"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Missing) != 1 || res.Missing[0] != "NOPE" {
		t.Errorf("expected NOPE in missing, got %v", res.Missing)
	}
}

func TestUnset_CustomDst(t *testing.T) {
	cfg, src := setupUnsetVault(t)
	writeUnsetEnv(t, src, "A=1\nB=2\n")
	dst := src + ".out"

	_, err := Unset(cfg, UnsetOptions{Src: src, Dst: dst, Keys: []string{"A"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if strings.Contains(string(data), "A=") {
		t.Error("key A should not appear in dst")
	}
	if !strings.Contains(string(data), "B=2") {
		t.Error("key B should be preserved in dst")
	}
	// source unchanged
	orig, _ := os.ReadFile(src)
	if !strings.Contains(string(orig), "A=1") {
		t.Error("source file should be unchanged")
	}
}

func TestUnset_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Unset(cfg, UnsetOptions{Src: filepath.Join(dir, ".env"), Keys: []string{"X"}})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestUnset_NoKeys(t *testing.T) {
	cfg, src := setupUnsetVault(t)
	writeUnsetEnv(t, src, "FOO=bar\n")
	_, err := Unset(cfg, UnsetOptions{Src: src, Keys: nil})
	if err == nil {
		t.Fatal("expected error when no keys provided")
	}
}
