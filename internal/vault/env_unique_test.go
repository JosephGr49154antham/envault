package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupUniqueVault(t *testing.T) (Config, string) {
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

func writeUniqueEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("writeUniqueEnv: %v", err)
	}
	return p
}

func TestUnique_RemovesDuplicates(t *testing.T) {
	cfg, dir := setupUniqueVault(t)
	src := writeUniqueEnv(t, dir, ".env", "# header\nKEY_A=first\nKEY_B=hello\nKEY_A=second\n")

	res, err := Unique(cfg, src, "")
	if err != nil {
		t.Fatalf("Unique: %v", err)
	}
	if len(res.Removed) != 1 || res.Removed[0] != "KEY_A" {
		t.Errorf("expected [KEY_A] removed, got %v", res.Removed)
	}

	out, _ := os.ReadFile(src)
	content := string(out)
	if count := countOccurrences(content, "KEY_A"); count != 1 {
		t.Errorf("expected 1 KEY_A in output, got %d", count)
	}
	if !contains(content, "KEY_A=second") {
		t.Error("expected last value (second) to be kept")
	}
}

func TestUnique_NoDuplicates(t *testing.T) {
	cfg, dir := setupUniqueVault(t)
	src := writeUniqueEnv(t, dir, ".env", "KEY_A=1\nKEY_B=2\nKEY_C=3\n")

	res, err := Unique(cfg, src, "")
	if err != nil {
		t.Fatalf("Unique: %v", err)
	}
	if len(res.Removed) != 0 {
		t.Errorf("expected no removals, got %v", res.Removed)
	}
	if res.Kept != 3 {
		t.Errorf("expected 3 kept lines, got %d", res.Kept)
	}
}

func TestUnique_CustomDst(t *testing.T) {
	cfg, dir := setupUniqueVault(t)
	src := writeUniqueEnv(t, dir, ".env", "X=1\nX=2\n")
	dst := filepath.Join(dir, ".env.unique")

	res, err := Unique(cfg, src, dst)
	if err != nil {
		t.Fatalf("Unique: %v", err)
	}
	if res.OutputPath != dst {
		t.Errorf("OutputPath = %q, want %q", res.OutputPath, dst)
	}
	// src should be untouched
	orig, _ := os.ReadFile(src)
	if countOccurrences(string(orig), "X=") != 2 {
		t.Error("source file should be unchanged")
	}
}

func TestUnique_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Unique(cfg, filepath.Join(dir, ".env"), "")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestUnique_PreservesComments(t *testing.T) {
	cfg, dir := setupUniqueVault(t)
	src := writeUniqueEnv(t, dir, ".env", "# comment\n\nKEY=val\n# another\nKEY=val2\n")

	_, err := Unique(cfg, src, "")
	if err != nil {
		t.Fatalf("Unique: %v", err)
	}
	out, _ := os.ReadFile(src)
	if !contains(string(out), "# comment") {
		t.Error("expected comments to be preserved")
	}
}

// helpers
func countOccurrences(s, sub string) int {
	count := 0
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			count++
		}
	}
	return count
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
