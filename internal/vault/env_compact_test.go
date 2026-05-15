package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCompactVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeCompactEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeCompactEnv: %v", err)
	}
}

func TestCompact_RemovesCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupCompactVault(t)
	src := filepath.Join(dir, ".env")
	writeCompactEnv(t, src, "# header comment\nFOO=bar\n\n# another comment\nBAZ=qux\n")

	dst := filepath.Join(dir, ".env.compact")
	err := Compact(cfg, CompactOptions{Src: src, Dst: dst, Overwrite: false})
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}

	data, _ := os.ReadFile(dst)
	result := string(data)
	if strings.Contains(result, "#") {
		t.Errorf("expected no comments, got:\n%s", result)
	}
	if strings.Contains(result, "\n\n") {
		t.Errorf("expected no blank lines, got:\n%s", result)
	}
	if !strings.Contains(result, "FOO=bar") || !strings.Contains(result, "BAZ=qux") {
		t.Errorf("expected key-value pairs preserved, got:\n%s", result)
	}
}

func TestCompact_DefaultDst(t *testing.T) {
	cfg, dir := setupCompactVault(t)
	src := filepath.Join(dir, ".env")
	original := "# comment\nKEY=val\n\n"
	writeCompactEnv(t, src, original)

	err := Compact(cfg, CompactOptions{Src: src})
	if err != nil {
		t.Fatalf("Compact in-place: %v", err)
	}

	data, _ := os.ReadFile(src)
	result := string(data)
	if strings.Contains(result, "#") {
		t.Errorf("expected comment removed, got: %s", result)
	}
	if !strings.Contains(result, "KEY=val") {
		t.Errorf("expected KEY=val preserved, got: %s", result)
	}
}

func TestCompact_NoOverwrite(t *testing.T) {
	cfg, dir := setupCompactVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeCompactEnv(t, src, "A=1\n")
	writeCompactEnv(t, dst, "existing\n")

	err := Compact(cfg, CompactOptions{Src: src, Dst: dst, Overwrite: false})
	if err == nil {
		t.Fatal("expected error when dst exists and overwrite is false")
	}
}

func TestCompact_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Compact(cfg, CompactOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
