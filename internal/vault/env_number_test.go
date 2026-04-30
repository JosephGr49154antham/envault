package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupNumberVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeNumberEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestNumber_AddsLineNumbers(t *testing.T) {
	cfg, dir := setupNumberVault(t)
	src := filepath.Join(dir, ".env")
	writeNumberEnv(t, src, "FOO=bar\nBAZ=qux\n")

	dst := filepath.Join(dir, ".env.numbered")
	err := Number(cfg, NumberOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("Number: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "   1  FOO=bar") {
		t.Errorf("expected line 1, got:\n%s", out)
	}
	if !strings.Contains(out, "   2  BAZ=qux") {
		t.Errorf("expected line 2, got:\n%s", out)
	}
}

func TestNumber_DefaultDst(t *testing.T) {
	cfg, dir := setupNumberVault(t)
	src := filepath.Join(dir, ".env")
	writeNumberEnv(t, src, "A=1\n")

	if err := Number(cfg, NumberOptions{Src: src}); err != nil {
		t.Fatalf("Number: %v", err)
	}

	expected := src + ".numbered"
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected default dst %s to exist", expected)
	}
}

func TestNumber_OnlyKV_SkipsComments(t *testing.T) {
	cfg, dir := setupNumberVault(t)
	src := filepath.Join(dir, ".env")
	writeNumberEnv(t, src, "# comment\nFOO=bar\n\nBAZ=qux\n")

	dst := filepath.Join(dir, ".env.numbered")
	err := Number(cfg, NumberOptions{Src: src, Dst: dst, OnlyKV: true})
	if err != nil {
		t.Fatalf("Number: %v", err)
	}

	data, _ := os.ReadFile(dst)
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	// comment line should have no leading number
	if strings.HasPrefix(strings.TrimSpace(lines[0]), "1") {
		t.Errorf("comment line should not be numbered, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "   1  FOO=bar") {
		t.Errorf("expected FOO=bar at line 1, got: %s", lines[1])
	}
}

func TestNumber_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
	}
	err := Number(cfg, NumberOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
