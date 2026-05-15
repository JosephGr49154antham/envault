package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envault/envault/internal/vault"
)

func setupDiffValuesVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := vault.Init(vault.Config{
		VaultDir:      filepath.Join(dir, ".vault"),
		RecipientsFile: filepath.Join(dir, ".vault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".vault", "env.age"),
	}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir
}

func writeDiffValuesEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDiffValues_IdenticalFiles(t *testing.T) {
	dir := setupDiffValuesVault(t)

	a := filepath.Join(dir, "a.env")
	b := filepath.Join(dir, "b.env")
	writeDiffValuesEnv(t, a, "FOO=bar\nBAZ=qux\n")
	writeDiffValuesEnv(t, b, "FOO=bar\nBAZ=qux\n")

	result, err := vault.DiffValues(a, b)
	if err != nil {
		t.Fatalf("DiffValues: %v", err)
	}
	if len(result.Changed) != 0 {
		t.Errorf("expected no changed keys, got %v", result.Changed)
	}
	if len(result.OnlyInA) != 0 {
		t.Errorf("expected no keys only in A, got %v", result.OnlyInA)
	}
	if len(result.OnlyInB) != 0 {
		t.Errorf("expected no keys only in B, got %v", result.OnlyInB)
	}
}

func TestDiffValues_ChangedValues(t *testing.T) {
	dir := setupDiffValuesVault(t)

	a := filepath.Join(dir, "a.env")
	b := filepath.Join(dir, "b.env")
	writeDiffValuesEnv(t, a, "FOO=old\nBAZ=same\n")
	writeDiffValuesEnv(t, b, "FOO=new\nBAZ=same\n")

	result, err := vault.DiffValues(a, b)
	if err != nil {
		t.Fatalf("DiffValues: %v", err)
	}
	if len(result.Changed) != 1 {
		t.Fatalf("expected 1 changed key, got %d: %v", len(result.Changed), result.Changed)
	}
	entry, ok := result.Changed["FOO"]
	if !ok {
		t.Fatal("expected FOO in changed")
	}
	if entry.A != "old" || entry.B != "new" {
		t.Errorf("unexpected values: A=%q B=%q", entry.A, entry.B)
	}
}

func TestDiffValues_OnlyInEachFile(t *testing.T) {
	dir := setupDiffValuesVault(t)

	a := filepath.Join(dir, "a.env")
	b := filepath.Join(dir, "b.env")
	writeDiffValuesEnv(t, a, "ONLY_A=1\nSHARED=x\n")
	writeDiffValuesEnv(t, b, "ONLY_B=2\nSHARED=x\n")

	result, err := vault.DiffValues(a, b)
	if err != nil {
		t.Fatalf("DiffValues: %v", err)
	}
	if len(result.OnlyInA) != 1 || result.OnlyInA[0] != "ONLY_A" {
		t.Errorf("expected [ONLY_A] in OnlyInA, got %v", result.OnlyInA)
	}
	if len(result.OnlyInB) != 1 || result.OnlyInB[0] != "ONLY_B" {
		t.Errorf("expected [ONLY_B] in OnlyInB, got %v", result.OnlyInB)
	}
}

func TestDiffValues_SkipsComments(t *testing.T) {
	dir := setupDiffValuesVault(t)

	a := filepath.Join(dir, "a.env")
	b := filepath.Join(dir, "b.env")
	writeDiffValuesEnv(t, a, "# comment\nFOO=bar\n")
	writeDiffValuesEnv(t, b, "# different comment\nFOO=bar\n")

	result, err := vault.DiffValues(a, b)
	if err != nil {
		t.Fatalf("DiffValues: %v", err)
	}
	if len(result.Changed) != 0 || len(result.OnlyInA) != 0 || len(result.OnlyInB) != 0 {
		t.Errorf("expected no diff ignoring comments, got changed=%v onlyA=%v onlyB=%v",
			result.Changed, result.OnlyInA, result.OnlyInB)
	}
}

func TestDiffValues_FileNotFound(t *testing.T) {
	_, err := vault.DiffValues("/nonexistent/a.env", "/nonexistent/b.env")
	if err == nil {
		t.Fatal("expected error for missing files")
	}
}
