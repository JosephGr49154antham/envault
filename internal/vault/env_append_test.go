package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupAppendVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := os.MkdirAll(cfg.VaultDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return cfg, dir
}

func writeAppendEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAppend_AddsNewKeys(t *testing.T) {
	cfg, dir := setupAppendVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeAppendEnv(t, src, "NEW_KEY=hello\nANOTHER=world\n")
	writeAppendEnv(t, dst, "EXISTING=yes\n")

	appended, err := Append(cfg, AppendOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(appended) != 2 {
		t.Fatalf("expected 2 appended keys, got %d", len(appended))
	}

	data, _ := os.ReadFile(dst)
	content := string(data)
	if !contains(content, "NEW_KEY=hello") || !contains(content, "ANOTHER=world") {
		t.Errorf("destination missing expected keys: %s", content)
	}
	if !contains(content, "EXISTING=yes") {
		t.Errorf("destination lost existing key")
	}
}

func TestAppend_SkipsExistingKeys(t *testing.T) {
	cfg, dir := setupAppendVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeAppendEnv(t, src, "KEY=new_value\n")
	writeAppendEnv(t, dst, "KEY=old_value\n")

	appended, err := Append(cfg, AppendOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(appended) != 0 {
		t.Errorf("expected 0 appended, got %d", len(appended))
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "KEY=old_value\n" {
		t.Errorf("destination was unexpectedly modified: %s", string(data))
	}
}

func TestAppend_FilterKeys(t *testing.T) {
	cfg, dir := setupAppendVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeAppendEnv(t, src, "ALPHA=1\nBETA=2\nGAMMA=3\n")
	writeAppendEnv(t, dst, "")

	appended, err := Append(cfg, AppendOptions{Src: src, Dst: dst, Keys: []string{"ALPHA", "GAMMA"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(appended) != 2 {
		t.Fatalf("expected 2 appended, got %d", len(appended))
	}

	data, _ := os.ReadFile(dst)
	if contains(string(data), "BETA") {
		t.Errorf("BETA should not have been appended")
	}
}

func TestAppend_DryRun(t *testing.T) {
	cfg, dir := setupAppendVault(t)
	src := filepath.Join(dir, "src.env")
	dst := filepath.Join(dir, "dst.env")
	writeAppendEnv(t, src, "NEW=1\n")
	writeAppendEnv(t, dst, "OLD=2\n")

	original, _ := os.ReadFile(dst)
	appended, err := Append(cfg, AppendOptions{Src: src, Dst: dst, DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(appended) != 1 {
		t.Errorf("expected 1 reported key, got %d", len(appended))
	}
	after, _ := os.ReadFile(dst)
	if string(after) != string(original) {
		t.Errorf("dry-run modified the file")
	}
}

func TestAppend_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{VaultDir: filepath.Join(dir, ".envault")}
	_, err := Append(cfg, AppendOptions{Src: "a.env", Dst: "b.env"})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
