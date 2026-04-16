package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTailVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncFile:     filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeTailEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestTail_DefaultN(t *testing.T) {
	cfg, _ := setupTailVault(t)
	var lines string
	for i := 1; i <= 15; i++ {
		lines += "KEY" + string(rune('A'+i-1)) + "=val\n"
	}
	writeTailEnv(t, cfg.PlainFile, lines)

	result, err := Tail(cfg, TailOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 10 {
		t.Errorf("expected 10 entries, got %d", len(result))
	}
}

func TestTail_CustomN(t *testing.T) {
	cfg, _ := setupTailVault(t)
	writeTailEnv(t, cfg.PlainFile, "A=1\nB=2\nC=3\nD=4\nE=5\n")

	result, err := Tail(cfg, TailOptions{N: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
	if result[0] != "C=3" {
		t.Errorf("expected C=3, got %s", result[0])
	}
}

func TestTail_KeysOnly(t *testing.T) {
	cfg, _ := setupTailVault(t)
	writeTailEnv(t, cfg.PlainFile, "FOO=bar\nBAZ=qux\nHELLO=world\n")

	result, err := Tail(cfg, TailOptions{N: 2, KeysOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0] != "BAZ" || result[1] != "HELLO" {
		t.Errorf("unexpected keys: %v", result)
	}
}

func TestTail_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	_, err := Tail(cfg, TailOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestTail_NGreaterThanEntries(t *testing.T) {
	cfg, _ := setupTailVault(t)
	writeTailEnv(t, cfg.PlainFile, "X=1\nY=2\n")

	result, err := Tail(cfg, TailOptions{N: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}
