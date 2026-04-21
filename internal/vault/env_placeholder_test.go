package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupPlaceholderVault(t *testing.T) (Config, string) {
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

func writePlaceholderEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestGeneratePlaceholder_CreatesFile(t *testing.T) {
	cfg, dir := setupPlaceholderVault(t)
	src := filepath.Join(dir, ".env")
	writePlaceholderEnv(t, src, "DB_HOST=localhost\nDB_PASS=secret\n")

	dst, err := GeneratePlaceholder(cfg, PlaceholderOptions{Src: src})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "DB_HOST=<DB_HOST>") {
		t.Errorf("expected placeholder for DB_HOST, got:\n%s", body)
	}
	if !strings.Contains(body, "DB_PASS=<DB_PASS>") {
		t.Errorf("expected placeholder for DB_PASS, got:\n%s", body)
	}
}

func TestGeneratePlaceholder_PreservesCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupPlaceholderVault(t)
	src := filepath.Join(dir, ".env")
	writePlaceholderEnv(t, src, "# header\n\nAPI_KEY=abc123\n")

	dst, err := GeneratePlaceholder(cfg, PlaceholderOptions{Src: src})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	body := string(data)
	if !strings.Contains(body, "# header") {
		t.Errorf("comment not preserved: %s", body)
	}
	if !strings.Contains(body, "API_KEY=<API_KEY>") {
		t.Errorf("key not placeholdered: %s", body)
	}
}

func TestGeneratePlaceholder_NoOverwrite(t *testing.T) {
	cfg, dir := setupPlaceholderVault(t)
	src := filepath.Join(dir, ".env")
	writePlaceholderEnv(t, src, "KEY=val\n")
	dst := filepath.Join(dir, ".env.placeholder")
	writePlaceholderEnv(t, dst, "existing\n")

	_, err := GeneratePlaceholder(cfg, PlaceholderOptions{Src: src, Dst: dst})
	if err == nil {
		t.Fatal("expected error when dst exists without overwrite")
	}
}

func TestGeneratePlaceholder_Overwrite(t *testing.T) {
	cfg, dir := setupPlaceholderVault(t)
	src := filepath.Join(dir, ".env")
	writePlaceholderEnv(t, src, "TOKEN=secret\n")
	dst := filepath.Join(dir, "out.env")
	writePlaceholderEnv(t, dst, "old content\n")

	_, err := GeneratePlaceholder(cfg, PlaceholderOptions{Src: src, Dst: dst, Overwrite: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if strings.Contains(string(data), "old content") {
		t.Error("old content should have been overwritten")
	}
}

func TestGeneratePlaceholder_CustomValueFmt(t *testing.T) {
	cfg, dir := setupPlaceholderVault(t)
	src := filepath.Join(dir, ".env")
	writePlaceholderEnv(t, src, "PORT=8080\n")

	dst, err := GeneratePlaceholder(cfg, PlaceholderOptions{
		Src:      src,
		ValueFmt: "CHANGE_ME_%s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "PORT=CHANGE_ME_PORT") {
		t.Errorf("custom format not applied: %s", string(data))
	}
}

func TestGeneratePlaceholder_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := GeneratePlaceholder(cfg, PlaceholderOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
