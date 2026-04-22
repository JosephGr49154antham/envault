package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupFlattenVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeFlattenEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeFlattenEnv: %v", err)
	}
}

func TestFlatten_Uppercase(t *testing.T) {
	cfg, dir := setupFlattenVault(t)
	src := filepath.Join(dir, ".env")
	writeFlattenEnv(t, src, "db_host=localhost\ndb_port=5432\n")

	dst, err := Flatten(cfg, src, FlattenOptions{Uppercase: true, Overwrite: false})
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "DB_HOST=localhost") {
		t.Errorf("expected uppercased key, got:\n%s", data)
	}
	if !strings.Contains(string(data), "DB_PORT=5432") {
		t.Errorf("expected uppercased key, got:\n%s", data)
	}
}

func TestFlatten_Prefix(t *testing.T) {
	cfg, dir := setupFlattenVault(t)
	src := filepath.Join(dir, ".env")
	writeFlattenEnv(t, src, "HOST=localhost\nPORT=5432\n")

	dst, err := Flatten(cfg, src, FlattenOptions{Prefix: "APP_", Overwrite: false})
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "APP_HOST=localhost") {
		t.Errorf("expected prefixed key, got:\n%s", data)
	}
}

func TestFlatten_DefaultDst(t *testing.T) {
	cfg, dir := setupFlattenVault(t)
	src := filepath.Join(dir, "secrets.env")
	writeFlattenEnv(t, src, "KEY=value\n")

	dst, err := Flatten(cfg, src, FlattenOptions{})
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	expected := filepath.Join(dir, "secrets.flat.env")
	if dst != expected {
		t.Errorf("expected dst %q, got %q", expected, dst)
	}
}

func TestFlatten_NoOverwrite(t *testing.T) {
	cfg, dir := setupFlattenVault(t)
	src := filepath.Join(dir, ".env")
	writeFlattenEnv(t, src, "KEY=value\n")
	dst := filepath.Join(dir, ".flat.env")
	writeFlattenEnv(t, dst, "existing=true\n")

	_, err := Flatten(cfg, src, FlattenOptions{Dst: dst, Overwrite: false})
	if err == nil {
		t.Fatal("expected error when destination exists")
	}
}

func TestFlatten_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{VaultDir: filepath.Join(dir, ".envault")}
	_, err := Flatten(cfg, filepath.Join(dir, ".env"), FlattenOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestFlatten_PreservesComments(t *testing.T) {
	cfg, dir := setupFlattenVault(t)
	src := filepath.Join(dir, ".env")
	writeFlattenEnv(t, src, "# header\nKEY=value\n\n# footer\n")

	dst, err := Flatten(cfg, src, FlattenOptions{Uppercase: true})
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "# header") {
		t.Errorf("expected comment preserved, got:\n%s", data)
	}
}
