package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupDedupVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeDedupEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing env file: %v", err)
	}
}

func TestDedup_RemovesDuplicates(t *testing.T) {
	cfg, dir := setupDedupVault(t)
	src := filepath.Join(dir, ".env")
	writeDedupEnv(t, src, "FOO=first\nBAR=one\nFOO=second\nBAZ=three\n")

	removed, err := Dedup(cfg, DedupOptions{Src: src})
	if err != nil {
		t.Fatalf("Dedup: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	got, _ := os.ReadFile(src)
	if string(got) != "FOO=first\nBAR=one\nBAZ=three\n" {
		t.Errorf("unexpected content: %q", string(got))
	}
}

func TestDedup_KeepLast(t *testing.T) {
	cfg, dir := setupDedupVault(t)
	src := filepath.Join(dir, ".env")
	writeDedupEnv(t, src, "FOO=first\nBAR=one\nFOO=second\n")

	_, err := Dedup(cfg, DedupOptions{Src: src, KeepLast: true})
	if err != nil {
		t.Fatalf("Dedup: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != "FOO=second\nBAR=one\n" {
		t.Errorf("unexpected content: %q", string(got))
	}
}

func TestDedup_NoDuplicates(t *testing.T) {
	cfg, dir := setupDedupVault(t)
	src := filepath.Join(dir, ".env")
	writeDedupEnv(t, src, "FOO=bar\nBAZ=qux\n")

	removed, err := Dedup(cfg, DedupOptions{Src: src})
	if err != nil {
		t.Fatalf("Dedup: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestDedup_CustomDst(t *testing.T) {
	cfg, dir := setupDedupVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.dedup")
	writeDedupEnv(t, src, "KEY=a\nKEY=b\n")

	_, err := Dedup(cfg, DedupOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("Dedup: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected dst to exist: %v", err)
	}
	original, _ := os.ReadFile(src)
	if string(original) != "KEY=a\nKEY=b\n" {
		t.Errorf("src should be unchanged, got %q", string(original))
	}
}

func TestDedup_NoOverwrite(t *testing.T) {
	cfg, dir := setupDedupVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeDedupEnv(t, src, "A=1\n")
	writeDedupEnv(t, dst, "existing\n")

	_, err := Dedup(cfg, DedupOptions{Src: src, Dst: dst, Overwrite: false})
	if err == nil {
		t.Error("expected error when dst exists without overwrite flag")
	}
}

func TestDedup_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile:  filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Dedup(cfg, DedupOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}
