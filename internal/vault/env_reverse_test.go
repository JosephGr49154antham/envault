package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupReverseVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeReverseEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestReverse_ReversesEntries(t *testing.T) {
	cfg, dir := setupReverseVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.reversed")
	writeReverseEnv(t, src, "A=1\nB=2\nC=3\n")

	err := Reverse(cfg, ReverseOptions{Src: src, Dst: dst, Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	lines := nonEmptyLines(string(data))
	if len(lines) != 3 || lines[0] != "C=3" || lines[1] != "B=2" || lines[2] != "A=1" {
		t.Fatalf("unexpected output: %q", lines)
	}
}

func TestReverse_DefaultDst(t *testing.T) {
	cfg, dir := setupReverseVault(t)
	src := filepath.Join(dir, ".env")
	writeReverseEnv(t, src, "X=10\nY=20\n")

	err := Reverse(cfg, ReverseOptions{Src: src})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := src + ".reversed"
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected dst %s to exist", expected)
	}
}

func TestReverse_NoOverwrite(t *testing.T) {
	cfg, dir := setupReverseVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeReverseEnv(t, src, "A=1\n")
	writeReverseEnv(t, dst, "existing")

	err := Reverse(cfg, ReverseOptions{Src: src, Dst: dst, Overwrite: false})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected overwrite error, got %v", err)
	}
}

func TestReverse_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
	}
	err := Reverse(cfg, ReverseOptions{Src: filepath.Join(dir, ".env")})
	if err == nil || !strings.Contains(err.Error(), "not initialised") {
		t.Fatalf("expected not-initialised error, got %v", err)
	}
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
