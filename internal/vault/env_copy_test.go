package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupCopyVault(t *testing.T) (Config, string) {
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

func writeCopyEnvFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCopy_AllKeys(t *testing.T) {
	cfg, dir := setupCopyVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeCopyEnvFile(t, src, "FOO=bar\nBAZ=qux\n")

	n, err := Copy(cfg, src, CopyOptions{Dst: dst})
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 keys copied, got %d", n)
	}
	m, _ := readEnvFileToMap(dst)
	if m["FOO"] != "bar" || m["BAZ"] != "qux" {
		t.Errorf("unexpected dst contents: %v", m)
	}
}

func TestCopy_SelectedKeys(t *testing.T) {
	cfg, dir := setupCopyVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeCopyEnvFile(t, src, "FOO=bar\nBAZ=qux\nSECRET=s3cr3t\n")

	n, err := Copy(cfg, src, CopyOptions{Dst: dst, Keys: []string{"FOO", "SECRET"}})
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 keys copied, got %d", n)
	}
	m, _ := readEnvFileToMap(dst)
	if _, ok := m["BAZ"]; ok {
		t.Error("BAZ should not have been copied")
	}
}

func TestCopy_NoOverwrite(t *testing.T) {
	cfg, dir := setupCopyVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeCopyEnvFile(t, src, "FOO=new\n")
	writeCopyEnvFile(t, dst, "FOO=old\n")

	n, err := Copy(cfg, src, CopyOptions{Dst: dst, Overwrite: false})
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 keys copied (no overwrite), got %d", n)
	}
	m, _ := readEnvFileToMap(dst)
	if m["FOO"] != "old" {
		t.Errorf("FOO should remain 'old', got %q", m["FOO"])
	}
}

func TestCopy_Overwrite(t *testing.T) {
	cfg, dir := setupCopyVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeCopyEnvFile(t, src, "FOO=new\n")
	writeCopyEnvFile(t, dst, "FOO=old\n")

	n, err := Copy(cfg, src, CopyOptions{Dst: dst, Overwrite: true})
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 key copied, got %d", n)
	}
	m, _ := readEnvFileToMap(dst)
	if m["FOO"] != "new" {
		t.Errorf("FOO should be 'new', got %q", m["FOO"])
	}
}

func TestCopy_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Copy(cfg, "src.env", CopyOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestCopy_MissingKey(t *testing.T) {
	cfg, dir := setupCopyVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeCopyEnvFile(t, src, "FOO=bar\n")

	_, err := Copy(cfg, src, CopyOptions{Dst: dst, Keys: []string{"MISSING"}})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}
