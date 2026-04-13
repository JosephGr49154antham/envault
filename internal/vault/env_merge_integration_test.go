package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMerge_EmptySrc verifies that merging an empty source leaves the base unchanged.
func TestMerge_EmptySrc(t *testing.T) {
	cfg, dir := setupMergeVault(t)
	base := filepath.Join(dir, ".env")
	src := filepath.Join(dir, ".env.empty")
	dst := filepath.Join(dir, ".env.out")

	writeEnvMerge(t, base, "X=1\nY=2\n")
	writeEnvMerge(t, src, "")

	res, err := Merge(cfg, base, src, dst)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(res.Added) != 0 || len(res.Updated) != 0 {
		t.Errorf("expected no changes, got added=%v updated=%v", res.Added, res.Updated)
	}
}

// TestMerge_EmptyBase verifies that merging into an empty base adds all src keys.
func TestMerge_EmptyBase(t *testing.T) {
	cfg, dir := setupMergeVault(t)
	base := filepath.Join(dir, ".env")
	src := filepath.Join(dir, ".env.src")
	dst := filepath.Join(dir, ".env.out")

	writeEnvMerge(t, base, "")
	writeEnvMerge(t, src, "A=alpha\nB=beta\n")

	res, err := Merge(cfg, base, src, dst)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if len(res.Added) != 2 {
		t.Errorf("expected 2 added keys, got %d: %v", len(res.Added), res.Added)
	}
	if len(res.Updated) != 0 {
		t.Errorf("expected 0 updated keys, got %v", res.Updated)
	}
}

// TestMerge_OutputFileCreated ensures the destination file is created.
func TestMerge_OutputFileCreated(t *testing.T) {
	cfg, dir := setupMergeVault(t)
	base := filepath.Join(dir, ".env")
	src := filepath.Join(dir, ".env.src")
	dst := filepath.Join(dir, "subdir", ".env.merged")

	_ = os.MkdirAll(filepath.Dir(dst), 0755)
	writeEnvMerge(t, base, "K=v\n")
	writeEnvMerge(t, src, "K=v\n")

	_, err := Merge(cfg, base, src, dst)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Errorf("expected output file to exist at %s", dst)
	}
}
