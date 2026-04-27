package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLower_MixedCaseValues(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env",
		"GREETING=Hello_World\nFLAG=TRUE\nNUM=42\n")
	dst := filepath.Join(dir, ".env.out")

	if err := Lower(src, dst, nil, false); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "GREETING=hello_world") {
		t.Errorf("expected GREETING=hello_world, got: %s", out)
	}
	if !strings.Contains(out, "FLAG=true") {
		t.Errorf("expected FLAG=true, got: %s", out)
	}
	// numeric values should be unchanged
	if !strings.Contains(out, "NUM=42") {
		t.Errorf("expected NUM=42 unchanged, got: %s", out)
	}
}

func TestLower_EmptyFile(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "")
	dst := filepath.Join(dir, ".env.out")

	if err := Lower(src, dst, nil, false); err != nil {
		t.Fatalf("Lower on empty file: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if len(strings.TrimSpace(string(data))) != 0 {
		t.Errorf("expected empty output, got: %q", string(data))
	}
}

func TestLower_DefaultDstPath(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "KEY=VAL\n")
	// passing src as dst mirrors the "default" path used by the CLI
	if err := Lower(src, src+".lower", nil, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src + ".lower"); err != nil {
		t.Errorf("expected default dst to exist")
	}
}
