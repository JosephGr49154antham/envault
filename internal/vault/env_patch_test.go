package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupPatchVault(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		PlainFile:     filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg
}

func writePatchEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestPatch_UpsertsNewKey(t *testing.T) {
	cfg := setupPatchVault(t)
	writePatchEnv(t, cfg.PlainFile, "FOO=bar\nBAZ=qux\n")

	err := Patch(cfg, PatchOptions{
		Upserts: map[string]string{"NEW_KEY": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPatchFileKeys(t, cfg.PlainFile)
	if _, ok := got["NEW_KEY"]; !ok {
		t.Error("expected NEW_KEY to be present")
	}
	if got["NEW_KEY"] != "hello" {
		t.Errorf("NEW_KEY = %q, want %q", got["NEW_KEY"], "hello")
	}
}

func TestPatch_UpdatesExistingKey(t *testing.T) {
	cfg := setupPatchVault(t)
	writePatchEnv(t, cfg.PlainFile, "FOO=oldval\n")

	err := Patch(cfg, PatchOptions{
		Upserts: map[string]string{"FOO": "newval"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPatchFileKeys(t, cfg.PlainFile)
	if got["FOO"] != "newval" {
		t.Errorf("FOO = %q, want %q", got["FOO"], "newval")
	}
}

func TestPatch_DeletesKey(t *testing.T) {
	cfg := setupPatchVault(t)
	writePatchEnv(t, cfg.PlainFile, "FOO=bar\nREMOVE_ME=gone\nBAZ=qux\n")

	err := Patch(cfg, PatchOptions{
		Deletions: []string{"REMOVE_ME"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPatchFileKeys(t, cfg.PlainFile)
	if _, ok := got["REMOVE_ME"]; ok {
		t.Error("expected REMOVE_ME to be deleted")
	}
	if _, ok := got["FOO"]; !ok {
		t.Error("expected FOO to be preserved")
	}
}

func TestPatch_NotInitialised(t *testing.T) {
	cfg := Config{
		VaultDir:  t.TempDir() + "/missing",
		PlainFile: t.TempDir() + "/.env",
	}
	err := Patch(cfg, PatchOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestPatch_CustomDst(t *testing.T) {
	cfg := setupPatchVault(t)
	writePatchEnv(t, cfg.PlainFile, "A=1\n")
	dst := filepath.Join(filepath.Dir(cfg.PlainFile), ".env.patched")

	err := Patch(cfg, PatchOptions{
		Src:     cfg.PlainFile,
		Dst:     dst,
		Upserts: map[string]string{"B": "2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readPatchFileKeys(t, dst)
	if got["B"] != "2" {
		t.Errorf("B = %q, want %q", got["B"], "2")
	}
	if got["A"] != "1" {
		t.Errorf("A = %q, want %q", got["A"], "1")
	}
}

// readPatchFileKeys reads key=value pairs from an env file into a map.
func readPatchFileKeys(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := make(map[string]string)
	for _, line := range splitLines(string(data)) {
		if line == "" || line[0] == '#' {
			continue
		}
		parts := splitOnFirst(line, '=')
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
