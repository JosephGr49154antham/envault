package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupUpperVault(t *testing.T) (Config, string) {
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

func writeUpperEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeUpperEnv: %v", err)
	}
}

func TestUpper_AllKeys(t *testing.T) {
	cfg, dir := setupUpperVault(t)
	src := filepath.Join(dir, ".env")
	writeUpperEnv(t, src, "db_host=localhost\napi_key=secret\n")

	if err := Upper(cfg, UpperOptions{Src: src}); err != nil {
		t.Fatalf("Upper: %v", err)
	}

	got, _ := os.ReadFile(src)
	if !strings.Contains(string(got), "DB_HOST=localhost") {
		t.Errorf("expected DB_HOST, got:\n%s", got)
	}
	if !strings.Contains(string(got), "API_KEY=secret") {
		t.Errorf("expected API_KEY, got:\n%s", got)
	}
}

func TestUpper_SelectedKeys(t *testing.T) {
	cfg, dir := setupUpperVault(t)
	src := filepath.Join(dir, ".env")
	writeUpperEnv(t, src, "db_host=localhost\napi_key=secret\n")

	if err := Upper(cfg, UpperOptions{Src: src, Keys: []string{"db_host"}}); err != nil {
		t.Fatalf("Upper: %v", err)
	}

	got, _ := os.ReadFile(src)
	if !strings.Contains(string(got), "DB_HOST=localhost") {
		t.Errorf("expected DB_HOST uppercased")
	}
	if !strings.Contains(string(got), "api_key=secret") {
		t.Errorf("expected api_key unchanged")
	}
}

func TestUpper_CustomDst(t *testing.T) {
	cfg, dir := setupUpperVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.upper")
	writeUpperEnv(t, src, "my_var=hello\n")

	if err := Upper(cfg, UpperOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("Upper: %v", err)
	}

	got, _ := os.ReadFile(dst)
	if !strings.Contains(string(got), "MY_VAR=hello") {
		t.Errorf("expected MY_VAR in dst, got:\n%s", got)
	}
	// src should be unchanged
	orig, _ := os.ReadFile(src)
	if !strings.Contains(string(orig), "my_var=hello") {
		t.Errorf("expected src unchanged, got:\n%s", orig)
	}
}

func TestUpper_PreservesCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupUpperVault(t)
	src := filepath.Join(dir, ".env")
	writeUpperEnv(t, src, "# header\n\nsome_key=val\n")

	if err := Upper(cfg, UpperOptions{Src: src}); err != nil {
		t.Fatalf("Upper: %v", err)
	}

	got, _ := os.ReadFile(src)
	if !strings.Contains(string(got), "# header") {
		t.Errorf("comment should be preserved")
	}
	if !strings.Contains(string(got), "SOME_KEY=val") {
		t.Errorf("key should be uppercased")
	}
}

func TestUpper_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Upper(cfg, UpperOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
