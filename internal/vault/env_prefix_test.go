package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupPrefixVault(t *testing.T) (Config, string) {
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

func writePrefixEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestPrefix_AddsPrefix(t *testing.T) {
	cfg, dir := setupPrefixVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.prefixed")
	writePrefixEnv(t, src, "DB_HOST=localhost\nDB_PORT=5432\n")

	err := Prefix(cfg, PrefixOptions{Src: src, Dst: dst, Prefix: "APP_", Overwrite: false})
	if err != nil {
		t.Fatalf("Prefix: %v", err)
	}

	data, _ := os.ReadFile(dst)
	content := string(data)
	if !contains(content, "APP_DB_HOST=localhost") {
		t.Errorf("expected APP_DB_HOST=localhost in output, got:\n%s", content)
	}
	if !contains(content, "APP_DB_PORT=5432") {
		t.Errorf("expected APP_DB_PORT=5432 in output, got:\n%s", content)
	}
}

func TestPrefix_RemovesPrefix(t *testing.T) {
	cfg, dir := setupPrefixVault(t)
	src := filepath.Join(dir, ".env")
	writePrefixEnv(t, src, "APP_HOST=localhost\nAPP_PORT=5432\nOTHER=val\n")

	err := Prefix(cfg, PrefixOptions{Src: src, Dst: src, Prefix: "APP_", Remove: true, Overwrite: true})
	if err != nil {
		t.Fatalf("Prefix remove: %v", err)
	}

	data, _ := os.ReadFile(src)
	content := string(data)
	if !contains(content, "HOST=localhost") {
		t.Errorf("expected HOST=localhost, got:\n%s", content)
	}
	if !contains(content, "OTHER=val") {
		t.Errorf("expected OTHER=val preserved, got:\n%s", content)
	}
}

func TestPrefix_NoOverwrite(t *testing.T) {
	cfg, dir := setupPrefixVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writePrefixEnv(t, src, "KEY=val\n")
	writePrefixEnv(t, dst, "existing\n")

	err := Prefix(cfg, PrefixOptions{Src: src, Dst: dst, Prefix: "X_", Overwrite: false})
	if err == nil {
		t.Fatal("expected error when dst exists and overwrite=false")
	}
}

func TestPrefix_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{VaultDir: filepath.Join(dir, ".envault")}
	err := Prefix(cfg, PrefixOptions{Src: ".env", Prefix: "X_"})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestPrefix_PreservesCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupPrefixVault(t)
	src := filepath.Join(dir, ".env")
	writePrefixEnv(t, src, "# comment\n\nKEY=val\n")

	err := Prefix(cfg, PrefixOptions{Src: src, Dst: src, Prefix: "P_", Overwrite: true})
	if err != nil {
		t.Fatalf("Prefix: %v", err)
	}

	data, _ := os.ReadFile(src)
	content := string(data)
	if !contains(content, "# comment") {
		t.Errorf("expected comment preserved, got:\n%s", content)
	}
	if !contains(content, "P_KEY=val") {
		t.Errorf("expected P_KEY=val, got:\n%s", content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
