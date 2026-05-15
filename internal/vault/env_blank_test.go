package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupBlankVault(t *testing.T) (Config, string) {
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

func writeBlankEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestBlank_AllKeys(t *testing.T) {
	cfg, dir := setupBlankVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.blanked")
	writeBlankEnv(t, src, "FOO=hello\nBAR=world\n# comment\n\nBAZ=123\n")

	if err := Blank(cfg, src, BlankOptions{Dst: dst, Overwrite: false}); err != nil {
		t.Fatalf("Blank: %v", err)
	}

	data, _ := os.ReadFile(dst)
	got := string(data)
	if strings.Contains(got, "hello") || strings.Contains(got, "world") || strings.Contains(got, "123") {
		t.Errorf("expected all values blanked, got:\n%s", got)
	}
	if !strings.Contains(got, "FOO=") || !strings.Contains(got, "BAR=") || !strings.Contains(got, "BAZ=") {
		t.Errorf("expected keys preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "# comment") {
		t.Errorf("expected comment preserved")
	}
}

func TestBlank_SelectedKeys(t *testing.T) {
	cfg, dir := setupBlankVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeBlankEnv(t, src, "SECRET=abc\nPUBLIC=open\n")

	if err := Blank(cfg, src, BlankOptions{Keys: []string{"SECRET"}, Dst: dst, Overwrite: false}); err != nil {
		t.Fatalf("Blank: %v", err)
	}

	data, _ := os.ReadFile(dst)
	got := string(data)
	if strings.Contains(got, "abc") {
		t.Errorf("SECRET value should be blanked")
	}
	if !strings.Contains(got, "PUBLIC=open") {
		t.Errorf("PUBLIC value should be preserved")
	}
}

func TestBlank_NoOverwrite(t *testing.T) {
	cfg, dir := setupBlankVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeBlankEnv(t, src, "A=1\n")
	writeBlankEnv(t, dst, "existing\n")

	err := Blank(cfg, src, BlankOptions{Dst: dst, Overwrite: false})
	if err == nil {
		t.Fatal("expected error for existing destination")
	}
}

func TestBlank_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Blank(cfg, filepath.Join(dir, ".env"), BlankOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
