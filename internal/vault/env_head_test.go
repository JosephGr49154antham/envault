package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupHeadVault(t *testing.T) (Config, string) {
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
	return cfg, dir
}

func writeHeadEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestHead_DefaultN(t *testing.T) {
	cfg, dir := setupHeadVault(t)
	src := filepath.Join(dir, ".env")
	var sb strings.Builder
	for i := 0; i < 15; i++ {
		sb.WriteString("KEY" + string(rune('A'+i)) + "=value\n")
	}
	writeHeadEnv(t, src, sb.String())

	out, err := Head(cfg, src, HeadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}
}

func TestHead_CustomN(t *testing.T) {
	cfg, dir := setupHeadVault(t)
	src := filepath.Join(dir, ".env")
	writeHeadEnv(t, src, "A=1\nB=2\nC=3\nD=4\nE=5\n")

	out, err := Head(cfg, src, HeadOptions{N: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "A=1" {
		t.Errorf("first line: got %q, want %q", lines[0], "A=1")
	}
}

func TestHead_KeysOnly(t *testing.T) {
	cfg, dir := setupHeadVault(t)
	src := filepath.Join(dir, ".env")
	writeHeadEnv(t, src, "# comment\nFOO=bar\n\nBAZ=qux\n")

	out, err := Head(cfg, src, HeadOptions{N: 10, KeysOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 key lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "FOO" || lines[1] != "BAZ" {
		t.Errorf("unexpected keys: %v", lines)
	}
}

func TestHead_WritesToDst(t *testing.T) {
	cfg, dir := setupHeadVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, "head.env")
	writeHeadEnv(t, src, "X=1\nY=2\n")

	result, err := Head(cfg, src, HeadOptions{N: 5, Dst: dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != dst {
		t.Errorf("expected result %q, got %q", dst, result)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "X=1") {
		t.Errorf("dst missing expected content, got: %s", data)
	}
}

func TestHead_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Head(cfg, filepath.Join(dir, ".env"), HeadOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
