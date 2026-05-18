package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTouchVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeTouchEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestTouch_AddsNewKeys(t *testing.T) {
	cfg, dir := setupTouchVault(t)
	src := filepath.Join(dir, ".env")
	writeTouchEnv(t, src, "EXISTING=hello\n")

	err := Touch(cfg, src, TouchOptions{Keys: []string{"NEW_KEY", "ANOTHER"}})
	if err != nil {
		t.Fatalf("Touch: %v", err)
	}

	data, _ := os.ReadFile(src)
	body := string(data)
	if !strings.Contains(body, "NEW_KEY=") {
		t.Errorf("expected NEW_KEY= in output, got:\n%s", body)
	}
	if !strings.Contains(body, "ANOTHER=") {
		t.Errorf("expected ANOTHER= in output, got:\n%s", body)
	}
	if !strings.Contains(body, "EXISTING=hello") {
		t.Errorf("expected EXISTING=hello preserved, got:\n%s", body)
	}
}

func TestTouch_SkipsExistingKeys(t *testing.T) {
	cfg, dir := setupTouchVault(t)
	src := filepath.Join(dir, ".env")
	writeTouchEnv(t, src, "FOO=bar\n")

	err := Touch(cfg, src, TouchOptions{Keys: []string{"FOO"}})
	if err != nil {
		t.Fatalf("Touch: %v", err)
	}

	data, _ := os.ReadFile(src)
	count := strings.Count(string(data), "FOO=")
	if count != 1 {
		t.Errorf("expected FOO= exactly once, got %d occurrences", count)
	}
}

func TestTouch_CustomDst(t *testing.T) {
	cfg, dir := setupTouchVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.touched")
	writeTouchEnv(t, src, "A=1\n")

	err := Touch(cfg, src, TouchOptions{Keys: []string{"B"}, Dst: dst, Overwrite: false})
	if err != nil {
		t.Fatalf("Touch: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "B=") {
		t.Errorf("expected B= in dst file")
	}
}

func TestTouch_NoOverwrite(t *testing.T) {
	cfg, dir := setupTouchVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeTouchEnv(t, src, "A=1\n")
	writeTouchEnv(t, dst, "existing\n")

	err := Touch(cfg, src, TouchOptions{Keys: []string{"B"}, Dst: dst, Overwrite: false})
	if err == nil {
		t.Fatal("expected error when dst exists and overwrite=false")
	}
}

func TestTouch_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	src := filepath.Join(dir, ".env")
	writeTouchEnv(t, src, "A=1\n")

	err := Touch(cfg, src, TouchOptions{Keys: []string{"B"}})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
