package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSortVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncFile:     filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeSortEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestSort_SortsKeysAlphabetically(t *testing.T) {
	cfg, dir := setupSortVault(t)
	src := filepath.Join(dir, ".env")
	writeSortEnv(t, src, "ZEBRA=1\nAPPLE=2\nMIDDLE=3\n")

	if err := Sort(cfg, src, ""); err != nil {
		t.Fatalf("Sort: %v", err)
	}

	data, _ := os.ReadFile(src)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if lines[0] != "APPLE=2" || lines[1] != "MIDDLE=3" || lines[2] != "ZEBRA=1" {
		t.Errorf("unexpected order: %v", lines)
	}
}

func TestSort_PreservesHeaderComments(t *testing.T) {
	cfg, dir := setupSortVault(t)
	src := filepath.Join(dir, ".env")
	writeSortEnv(t, src, "# top comment\n\nZOO=z\nANT=a\n")

	if err := Sort(cfg, src, ""); err != nil {
		t.Fatalf("Sort: %v", err)
	}

	data, _ := os.ReadFile(src)
	content := string(data)
	if !strings.HasPrefix(content, "# top comment\n") {
		t.Errorf("comment not preserved at top: %q", content)
	}
	if strings.Index(content, "ANT=a") > strings.Index(content, "ZOO=z") {
		t.Errorf("keys not sorted after header")
	}
}

func TestSort_CustomDst(t *testing.T) {
	cfg, dir := setupSortVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.sorted")
	writeSortEnv(t, src, "Z=last\nA=first\n")

	if err := Sort(cfg, src, dst); err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("dst not created: %v", err)
	}
	origData, _ := os.ReadFile(src)
	if strings.Contains(string(origData), "A=first\nZ=last") {
		t.Error("source file should not be modified when dst differs")
	}
}

func TestSort_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	err := Sort(cfg, "", "")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
