package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupUnquoteVault(t *testing.T) Config {
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
	return cfg
}

func writeUnquoteEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestUnquote_RemovesDoubleQuotes(t *testing.T) {
	cfg := setupUnquoteVault(t)
	src := filepath.Join(filepath.Dir(cfg.VaultDir), ".env")
	writeUnquoteEnv(t, src, `FOO="hello world"
BAR="simple"
`)

	if err := Unquote(cfg, src, src, true); err != nil {
		t.Fatalf("Unquote: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != "FOO=hello world\nBAR=simple\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestUnquote_RemovesSingleQuotes(t *testing.T) {
	cfg := setupUnquoteVault(t)
	src := filepath.Join(filepath.Dir(cfg.VaultDir), ".env")
	writeUnquoteEnv(t, src, "KEY='value'\n")

	if err := Unquote(cfg, src, src, true); err != nil {
		t.Fatalf("Unquote: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != "KEY=value\n" {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestUnquote_PreservesCommentsAndBlanks(t *testing.T) {
	cfg := setupUnquoteVault(t)
	src := filepath.Join(filepath.Dir(cfg.VaultDir), ".env")
	input := "# comment\n\nFOO=\"bar\"\n"
	writeUnquoteEnv(t, src, input)

	if err := Unquote(cfg, src, src, true); err != nil {
		t.Fatalf("Unquote: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != "# comment\n\nFOO=bar\n" {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestUnquote_NoOverwrite(t *testing.T) {
	cfg := setupUnquoteVault(t)
	base := filepath.Dir(cfg.VaultDir)
	src := filepath.Join(base, ".env")
	dst := filepath.Join(base, ".env.out")
	writeUnquoteEnv(t, src, "A=\"1\"\n")
	writeUnquoteEnv(t, dst, "existing")

	err := Unquote(cfg, src, dst, false)
	if err == nil {
		t.Fatal("expected error when dst exists and overwrite=false")
	}
}

func TestUnquote_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/missing"}
	err := Unquote(cfg, ".env", ".env", true)
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
