package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTrimVault(t *testing.T) (Config, string) {
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

func writeTrimEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestTrim_RemovesComments(t *testing.T) {
	cfg, dir := setupTrimVault(t)
	src := filepath.Join(dir, ".env")
	writeTrimEnv(t, src, "# comment\nKEY=val\n\nANOTHER=x\n")

	if err := Trim(cfg, src, TrimOptions{RemoveComments: true}); err != nil {
		t.Fatalf("Trim: %v", err)
	}
	b, _ := os.ReadFile(src)
	if strings.Contains(string(b), "# comment") {
		t.Error("expected comment to be removed")
	}
	if !strings.Contains(string(b), "KEY=val") {
		t.Error("expected KEY=val to be preserved")
	}
}

func TestTrim_RemovesBlanks(t *testing.T) {
	cfg, dir := setupTrimVault(t)
	src := filepath.Join(dir, ".env")
	writeTrimEnv(t, src, "KEY=val\n\n   \nANOTHER=x\n")

	if err := Trim(cfg, src, TrimOptions{RemoveBlanks: true}); err != nil {
		t.Fatalf("Trim: %v", err)
	}
	b, _ := os.ReadFile(src)
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			t.Errorf("unexpected blank line: %q", l)
		}
	}
}

func TestTrim_TrimValues(t *testing.T) {
	cfg, dir := setupTrimVault(t)
	src := filepath.Join(dir, ".env")
	writeTrimEnv(t, src, "KEY =  hello world  \nB= foo \n")

	if err := Trim(cfg, src, TrimOptions{TrimValues: true}); err != nil {
		t.Fatalf("Trim: %v", err)
	}
	b, _ := os.ReadFile(src)
	if !strings.Contains(string(b), "KEY=hello world") {
		t.Errorf("expected trimmed key=value, got: %s", string(b))
	}
	if !strings.Contains(string(b), "B=foo") {
		t.Errorf("expected B=foo, got: %s", string(b))
	}
}

func TestTrim_CustomDst(t *testing.T) {
	cfg, dir := setupTrimVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.trimmed")
	writeTrimEnv(t, src, "# comment\nKEY=val\n")

	if err := Trim(cfg, src, TrimOptions{RemoveComments: true, Dst: dst}); err != nil {
		t.Fatalf("Trim: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("expected dst file to exist: %v", err)
	}
	// src should be untouched
	b, _ := os.ReadFile(src)
	if !strings.Contains(string(b), "# comment") {
		t.Error("src should be unchanged when Dst is set")
	}
}

func TestTrim_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Trim(cfg, filepath.Join(dir, ".env"), TrimOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
