package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupDiffKeysVault(t *testing.T) (Config, string) {
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

func writeDiffEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

func TestDiffKeys_DisjointFiles(t *testing.T) {
	cfg, dir := setupDiffKeysVault(t)
	fileA := filepath.Join(dir, "a.env")
	fileB := filepath.Join(dir, "b.env")
	writeDiffEnv(t, fileA, "FOO=1\nBAR=2\n")
	writeDiffEnv(t, fileB, "BAZ=3\nQUX=4\n")

	res, err := DiffKeys(cfg, fileA, fileB)
	if err != nil {
		t.Fatalf("DiffKeys: %v", err)
	}
	if len(res.InBoth) != 0 {
		t.Errorf("expected no shared keys, got %v", res.InBoth)
	}
	if len(res.OnlyInA) != 2 {
		t.Errorf("expected 2 keys only in A, got %v", res.OnlyInA)
	}
	if len(res.OnlyInB) != 2 {
		t.Errorf("expected 2 keys only in B, got %v", res.OnlyInB)
	}
}

func TestDiffKeys_PartialOverlap(t *testing.T) {
	cfg, dir := setupDiffKeysVault(t)
	fileA := filepath.Join(dir, "a.env")
	fileB := filepath.Join(dir, "b.env")
	writeDiffEnv(t, fileA, "# comment\nFOO=1\nSHARED=x\n")
	writeDiffEnv(t, fileB, "SHARED=y\nBAR=2\n")

	res, err := DiffKeys(cfg, fileA, fileB)
	if err != nil {
		t.Fatalf("DiffKeys: %v", err)
	}
	if len(res.InBoth) != 1 || res.InBoth[0] != "SHARED" {
		t.Errorf("expected SHARED in both, got %v", res.InBoth)
	}
	if len(res.OnlyInA) != 1 || res.OnlyInA[0] != "FOO" {
		t.Errorf("expected FOO only in A, got %v", res.OnlyInA)
	}
	if len(res.OnlyInB) != 1 || res.OnlyInB[0] != "BAR" {
		t.Errorf("expected BAR only in B, got %v", res.OnlyInB)
	}
}

func TestDiffKeys_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := DiffKeys(cfg, "a.env", "b.env")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestDiffKeys_MissingFile(t *testing.T) {
	cfg, dir := setupDiffKeysVault(t)
	fileA := filepath.Join(dir, "a.env")
	writeDiffEnv(t, fileA, "FOO=1\n")

	_, err := DiffKeys(cfg, fileA, filepath.Join(dir, "missing.env"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
