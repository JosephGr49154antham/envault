package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupStripVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeStripEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestStrip_RemovesComments(t *testing.T) {
	cfg, dir := setupStripVault(t)
	src := filepath.Join(dir, ".env")
	writeStripEnv(t, src, "# comment\nKEY=value\n# another\nFOO=bar\n")

	n, err := Strip(cfg, src, StripOptions{RemoveComments: true})
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 removed, got %d", n)
	}
	data, _ := os.ReadFile(src)
	if strings.Contains(string(data), "#") {
		t.Error("comments should have been removed")
	}
}

func TestStrip_RemovesBlanks(t *testing.T) {
	cfg, dir := setupStripVault(t)
	src := filepath.Join(dir, ".env")
	writeStripEnv(t, src, "KEY=value\n\n\nFOO=bar\n   \n")

	n, err := Strip(cfg, src, StripOptions{RemoveBlanks: true})
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 removed, got %d", n)
	}
}

func TestStrip_CustomDst(t *testing.T) {
	cfg, dir := setupStripVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.stripped")
	writeStripEnv(t, src, "# header\nKEY=value\n")

	_, err := Strip(cfg, src, StripOptions{RemoveComments: true, Dst: dst})
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("dst file not created: %v", err)
	}
	// src should be unchanged
	orig, _ := os.ReadFile(src)
	if !strings.Contains(string(orig), "# header") {
		t.Error("src should not be modified when dst is set")
	}
}

func TestStrip_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Strip(cfg, filepath.Join(dir, ".env"), StripOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
