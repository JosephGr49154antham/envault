package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupMergeVault(t *testing.T) (Config, string) {
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

func writeEnvMerge(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeEnvMerge: %v", err)
	}
}

func TestMerge_AddedAndUpdated(t *testing.T) {
	cfg, dir := setupMergeVault(t)
	base := filepath.Join(dir, ".env")
	src := filepath.Join(dir, ".env.new")
	dst := filepath.Join(dir, ".env.merged")

	writeEnvMerge(t, base, "FOO=bar\nKEEP=same\n")
	writeEnvMerge(t, src, "FOO=baz\nNEW=hello\nKEEP=same\n")

	res, err := Merge(cfg, base, src, dst)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(res.Added) != 1 || res.Added[0] != "NEW" {
		t.Errorf("expected Added=[NEW], got %v", res.Added)
	}
	if len(res.Updated) != 1 || res.Updated[0] != "FOO" {
		t.Errorf("expected Updated=[FOO], got %v", res.Updated)
	}
	if len(res.Kept) != 1 || res.Kept[0] != "KEEP" {
		t.Errorf("expected Kept=[KEEP], got %v", res.Kept)
	}
	if res.Merged["FOO"] != "baz" {
		t.Errorf("expected FOO=baz, got %s", res.Merged["FOO"])
	}
	if res.Merged["NEW"] != "hello" {
		t.Errorf("expected NEW=hello, got %s", res.Merged["NEW"])
	}
}

func TestMerge_InPlace(t *testing.T) {
	cfg, dir := setupMergeVault(t)
	base := filepath.Join(dir, ".env")
	src := filepath.Join(dir, ".env.src")

	writeEnvMerge(t, base, "A=1\n")
	writeEnvMerge(t, src, "A=2\nB=3\n")

	_, err := Merge(cfg, base, src, "")
	if err != nil {
		t.Fatalf("Merge in-place: %v", err)
	}

	got, err := parseEnvFile(base)
	if err != nil {
		t.Fatalf("parseEnvFile: %v", err)
	}
	if got["A"] != "2" {
		t.Errorf("expected A=2, got %s", got["A"])
	}
	if got["B"] != "3" {
		t.Errorf("expected B=3, got %s", got["B"])
	}
}

func TestMerge_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Merge(cfg, ".env", ".env.src", "")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestMerge_MissingBase(t *testing.T) {
	cfg, dir := setupMergeVault(t)
	src := filepath.Join(dir, ".env.src")
	writeEnvMerge(t, src, "A=1\n")

	_, err := Merge(cfg, filepath.Join(dir, "nonexistent"), src, "")
	if err == nil {
		t.Fatal("expected error for missing base file")
	}
}
