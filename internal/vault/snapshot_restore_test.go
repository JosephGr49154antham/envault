package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupRestoreVault(t *testing.T) (Config, string) {
	t.Helper()
	tmp := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(tmp, ".envault"),
		PlainFile:     filepath.Join(tmp, ".env"),
		EncryptedFile: filepath.Join(tmp, ".envault", "secrets.env.age"),
		RecipientsFile: filepath.Join(tmp, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, tmp
}

func writeSnapshot(t *testing.T, cfg Config, label string, entries map[string]string) {
	t.Helper()
	dir := snapshotDir(cfg)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	snap := struct {
		Label   string            `json:"label"`
		Entries map[string]string `json:"entries"`
	}{Label: label, Entries: entries}
	data, _ := json.Marshal(snap)
	name := sanitiseLabel(label) + ".json"
	if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestRestoreSnapshot_WritesPlainFile(t *testing.T) {
	cfg, _ := setupRestoreVault(t)
	entries := map[string]string{"FOO": "bar", "BAZ": "qux"}
	writeSnapshot(t, cfg, "v1", entries)

	if err := RestoreSnapshot(cfg, "v1", ""); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}

	data, err := os.ReadFile(cfg.PlainFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	for k, v := range entries {
		expected := k + "=" + v
		if !contains(content, expected) {
			t.Errorf("expected %q in restored file, got:\n%s", expected, content)
		}
	}
}

func TestRestoreSnapshot_CustomDst(t *testing.T) {
	cfg, tmp := setupRestoreVault(t)
	writeSnapshot(t, cfg, "release", map[string]string{"KEY": "value"})

	dst := filepath.Join(tmp, ".env.restored")
	if err := RestoreSnapshot(cfg, "release", dst); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected restored file at %s: %v", dst, err)
	}
}

func TestRestoreSnapshot_NotFound(t *testing.T) {
	cfg, _ := setupRestoreVault(t)
	err := RestoreSnapshot(cfg, "nonexistent", "")
	if err == nil {
		t.Fatal("expected error for missing snapshot, got nil")
	}
}

func TestRestoreSnapshot_NotInitialised(t *testing.T) {
	tmp := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(tmp, ".envault"),
		PlainFile:     filepath.Join(tmp, ".env"),
		EncryptedFile: filepath.Join(tmp, ".envault", "secrets.env.age"),
		RecipientsFile: filepath.Join(tmp, ".envault", "recipients.txt"),
	}
	err := RestoreSnapshot(cfg, "v1", "")
	if err == nil {
		t.Fatal("expected error for uninitialised vault, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (
		len(substr) == 0 || func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}
