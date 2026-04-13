package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSplit_RemainderOnly verifies that when no prefixes are given but a
// remainder destination is set, every key lands in the remainder file.
func TestSplit_RemainderOnly(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	src := filepath.Join(dir, ".env")
	writeSplitEnv(t, src, "FOO=1\nBAR=2\n")

	out := filepath.Join(dir, "all.env")
	if err := Split(cfg, src, SplitOptions{Remainder: out}); err != nil {
		t.Fatalf("split: %v", err)
	}

	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "FOO=1") || !strings.Contains(string(data), "BAR=2") {
		t.Errorf("remainder file missing keys: %s", data)
	}
}

// TestSplit_UnmatchedKeyDroppedWithoutRemainder ensures that keys with no
// matching prefix and no Remainder destination are silently dropped.
func TestSplit_UnmatchedKeyDroppedWithoutRemainder(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	src := filepath.Join(dir, ".env")
	writeSplitEnv(t, src, "DB_HOST=localhost\nUNMATCHED=value\n")

	dbDest := filepath.Join(dir, "db.env")
	opts := SplitOptions{
		Prefixes: map[string]string{"DB_": dbDest},
	}
	if err := Split(cfg, src, opts); err != nil {
		t.Fatalf("split: %v", err)
	}

	data, _ := os.ReadFile(dbDest)
	if strings.Contains(string(data), "UNMATCHED") {
		t.Errorf("expected UNMATCHED to be dropped, but found it in db.env")
	}
}

// TestSplit_CreatesOutputDirectories verifies nested output directories are
// created automatically.
func TestSplit_CreatesOutputDirectories(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	src := filepath.Join(dir, ".env")
	writeSplitEnv(t, src, "DB_HOST=localhost\n")

	nestedDest := filepath.Join(dir, "nested", "deep", "db.env")
	opts := SplitOptions{
		Prefixes: map[string]string{"DB_": nestedDest},
	}
	if err := Split(cfg, src, opts); err != nil {
		t.Fatalf("split: %v", err)
	}

	if _, err := os.Stat(nestedDest); err != nil {
		t.Errorf("expected nested output file to exist: %v", err)
	}
}
