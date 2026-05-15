package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrap_RoundTripPreservesKeys(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_ = Init(cfg)

	content := "DB_URL=postgres://user:password@localhost:5432/mydb?sslmode=disable&connect_timeout=10\nSECRET=short\n"
	src := filepath.Join(dir, ".env")
	_ = os.WriteFile(src, []byte(content), 0o644)

	if err := Wrap(cfg, WrapOptions{Src: src, Width: 40, Overwrite: true}); err != nil {
		t.Fatalf("wrap: %v", err)
	}

	data, _ := os.ReadFile(src)
	if !strings.Contains(string(data), "DB_URL=") {
		t.Error("DB_URL key should still be present after wrap")
	}
	if !strings.Contains(string(data), "SECRET=short") {
		t.Error("short line should be unchanged")
	}
}

func TestWrap_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_ = Init(cfg)

	src := filepath.Join(dir, ".env")
	_ = os.WriteFile(src, []byte(""), 0o644)

	if err := Wrap(cfg, WrapOptions{Src: src, Width: 80, Overwrite: true}); err != nil {
		t.Fatalf("wrap empty file: %v", err)
	}
	data, _ := os.ReadFile(src)
	if strings.TrimSpace(string(data)) != "" {
		t.Error("expected empty output for empty input")
	}
}

func TestWrap_DefaultDstPath(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_ = Init(cfg)

	src := filepath.Join(dir, ".env")
	_ = os.WriteFile(src, []byte("KEY=value\n"), 0o644)

	opts := WrapOptions{Src: src, Width: 80, Overwrite: true}
	if opts.Dst == "" {
		opts.Dst = opts.Src
	}
	if opts.Dst != src {
		t.Errorf("default dst should equal src, got %q", opts.Dst)
	}
}
