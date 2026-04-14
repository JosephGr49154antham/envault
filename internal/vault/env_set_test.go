package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupSetVault(t *testing.T) (Config, string) {
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

func writeSetEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestSet_AddsNewKey(t *testing.T) {
	cfg, dir := setupSetVault(t)
	env := filepath.Join(dir, ".env")
	writeSetEnv(t, env, "EXISTING=yes\n")

	if err := Set(cfg, env, "", "NEW_KEY", "hello"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	data, _ := os.ReadFile(env)
	content := string(data)
	if !contains(content, "NEW_KEY=hello") {
		t.Errorf("expected NEW_KEY=hello in output, got:\n%s", content)
	}
	if !contains(content, "EXISTING=yes") {
		t.Errorf("expected EXISTING=yes preserved, got:\n%s", content)
	}
}

func TestSet_UpdatesExistingKey(t *testing.T) {
	cfg, dir := setupSetVault(t)
	env := filepath.Join(dir, ".env")
	writeSetEnv(t, env, "FOO=old\nBAR=keep\n")

	if err := Set(cfg, env, "", "FOO", "new"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	data, _ := os.ReadFile(env)
	content := string(data)
	if !contains(content, "FOO=new") {
		t.Errorf("expected FOO=new, got:\n%s", content)
	}
	if contains(content, "FOO=old") {
		t.Errorf("old value should be replaced, got:\n%s", content)
	}
}

func TestSet_CustomDst(t *testing.T) {
	cfg, dir := setupSetVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeSetEnv(t, src, "A=1\n")

	if err := Set(cfg, src, dst, "B", "2"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !contains(string(data), "B=2") {
		t.Errorf("expected B=2 in dst, got: %s", data)
	}
}

func TestSet_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/nope"}
	err := Set(cfg, ".env", "", "K", "v")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestUnset_RemovesKey(t *testing.T) {
	cfg, dir := setupSetVault(t)
	env := filepath.Join(dir, ".env")
	writeSetEnv(t, env, "REMOVE=me\nKEEP=this\n")

	if err := Unset(cfg, env, "", "REMOVE"); err != nil {
		t.Fatalf("Unset: %v", err)
	}

	data, _ := os.ReadFile(env)
	content := string(data)
	if contains(content, "REMOVE") {
		t.Errorf("expected REMOVE to be gone, got:\n%s", content)
	}
	if !contains(content, "KEEP=this") {
		t.Errorf("expected KEEP=this preserved, got:\n%s", content)
	}
}

func TestUnset_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/nope"}
	err := Unset(cfg, ".env", "", "K")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
