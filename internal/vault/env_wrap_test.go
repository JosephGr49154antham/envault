package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupWrapVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeWrapEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	return p
}

func TestWrap_ShortLinesUnchanged(t *testing.T) {
	cfg, dir := setupWrapVault(t)
	src := writeWrapEnv(t, dir, ".env", "KEY=short\nOTHER=value\n")
	err := Wrap(cfg, WrapOptions{Src: src, Width: 80, Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(src)
	if strings.Contains(string(data), "\\") {
		t.Error("short lines should not be wrapped")
	}
}

func TestWrap_LongValueIsFolded(t *testing.T) {
	cfg, dir := setupWrapVault(t)
	value := strings.Repeat("x", 120)
	src := writeWrapEnv(t, dir, ".env", "LONG="+value+"\n")
	err := Wrap(cfg, WrapOptions{Src: src, Width: 40, Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(src)
	if !strings.Contains(string(data), "\\") {
		t.Error("expected backslash continuation in output")
	}
}

func TestWrap_CommentsPreserved(t *testing.T) {
	cfg, dir := setupWrapVault(t)
	src := writeWrapEnv(t, dir, ".env", "# header\nKEY=val\n")
	if err := Wrap(cfg, WrapOptions{Src: src, Width: 80, Overwrite: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(src)
	if !strings.Contains(string(data), "# header") {
		t.Error("comment line should be preserved")
	}
}

func TestWrap_NoOverwrite(t *testing.T) {
	cfg, dir := setupWrapVault(t)
	src := writeWrapEnv(t, dir, ".env", "KEY=value\n")
	dst := filepath.Join(dir, "out.env")
	_ = os.WriteFile(dst, []byte("existing"), 0o644)
	err := Wrap(cfg, WrapOptions{Src: src, Dst: dst, Width: 80, Overwrite: false})
	if err == nil {
		t.Error("expected error when dst exists and overwrite is false")
	}
}

func TestWrap_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/nope"}
	err := Wrap(cfg, WrapOptions{Src: ".env", Width: 80})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestWrapValue_SplitsCorrectly(t *testing.T) {
	lines := wrapValue("KEY", strings.Repeat("a", 60), 20)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %d", len(lines))
	}
	for i, l := range lines[:len(lines)-1] {
		if !strings.HasSuffix(l, "\\") {
			t.Errorf("line %d missing continuation backslash: %q", i, l)
		}
	}
}
