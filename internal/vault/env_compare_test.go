package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupCompareVault(t *testing.T) (Config, string) {
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

func writeCompareEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestCompare_Identical(t *testing.T) {
	cfg, dir := setupCompareVault(t)
	a := filepath.Join(dir, ".env.a")
	b := filepath.Join(dir, ".env.b")
	content := "KEY1=value1\nKEY2=value2\n"
	writeCompareEnv(t, a, content)
	writeCompareEnv(t, b, content)

	res, err := Compare(cfg, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.HasDifferences() {
		t.Errorf("expected no differences, got %+v", res)
	}
	if len(res.Identical) != 2 {
		t.Errorf("expected 2 identical keys, got %d", len(res.Identical))
	}
}

func TestCompare_ChangedValues(t *testing.T) {
	cfg, dir := setupCompareVault(t)
	a := filepath.Join(dir, ".env.a")
	b := filepath.Join(dir, ".env.b")
	writeCompareEnv(t, a, "KEY=old\n")
	writeCompareEnv(t, b, "KEY=new\n")

	res, err := Compare(cfg, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasDifferences() {
		t.Error("expected differences")
	}
	if pair, ok := res.Changed["KEY"]; !ok || pair[0] != "old" || pair[1] != "new" {
		t.Errorf("unexpected changed entry: %v", res.Changed)
	}
}

func TestCompare_OnlyInEachFile(t *testing.T) {
	cfg, dir := setupCompareVault(t)
	a := filepath.Join(dir, ".env.a")
	b := filepath.Join(dir, ".env.b")
	writeCompareEnv(t, a, "ONLY_A=1\nSHARED=x\n")
	writeCompareEnv(t, b, "ONLY_B=2\nSHARED=x\n")

	res, err := Compare(cfg, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.OnlyInA) != 1 || res.OnlyInA[0] != "ONLY_A" {
		t.Errorf("expected ONLY_A in OnlyInA, got %v", res.OnlyInA)
	}
	if len(res.OnlyInB) != 1 || res.OnlyInB[0] != "ONLY_B" {
		t.Errorf("expected ONLY_B in OnlyInB, got %v", res.OnlyInB)
	}
}

func TestCompare_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Compare(cfg, "a", "b")
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestCompare_SkipsCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupCompareVault(t)
	a := filepath.Join(dir, ".env.a")
	b := filepath.Join(dir, ".env.b")
	writeCompareEnv(t, a, "# comment\n\nKEY=val\n")
	writeCompareEnv(t, b, "KEY=val\n")

	res, err := Compare(cfg, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.HasDifferences() {
		t.Errorf("expected no differences, got %+v", res)
	}
}
