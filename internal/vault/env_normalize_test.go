package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupNormalizeVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := Init(DefaultConfig(dir)); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir
}

func writeNormalizeEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestNormalize_TrimsAndQuotes(t *testing.T) {
	dir := setupNormalizeVault(t)
	src := filepath.Join(dir, ".env")
	writeNormalizeEnv(t, src, "KEY=  hello world  \nFOO=bar\n")

	dst := filepath.Join(dir, ".env.normalized")
	if err := Normalize(src, dst, false); err != nil {
		t.Fatalf("Normalize: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, `KEY="hello world"`) {
		t.Errorf("expected KEY to be trimmed and quoted, got: %s", out)
	}
	if !strings.Contains(out, "FOO=bar") {
		t.Errorf("expected FOO=bar to remain unchanged, got: %s", out)
	}
}

func TestNormalize_DefaultDst(t *testing.T) {
	dir := setupNormalizeVault(t)
	src := filepath.Join(dir, ".env")
	writeNormalizeEnv(t, src, "KEY=value\n")

	if err := Normalize(src, "", false); err != nil {
		t.Fatalf("Normalize: %v", err)
	}

	// default dst should be src (in-place)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "KEY=value") {
		t.Errorf("expected in-place normalised file, got: %s", string(data))
	}
}

func TestNormalize_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, ".env")
	writeNormalizeEnv(t, src, "KEY=value\n")

	err := Normalize(src, "", false)
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestNormalize_PreservesComments(t *testing.T) {
	dir := setupNormalizeVault(t)
	src := filepath.Join(dir, ".env")
	writeNormalizeEnv(t, src, "# header comment\nKEY=value\n\n# another comment\nFOO=bar\n")

	dst := filepath.Join(dir, ".env.out")
	if err := Normalize(src, dst, false); err != nil {
		t.Fatalf("Normalize: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "# header comment") {
		t.Errorf("expected header comment to be preserved, got: %s", out)
	}
	if !strings.Contains(out, "# another comment") {
		t.Errorf("expected inline comment to be preserved, got: %s", out)
	}
}

func TestNormalize_UppercaseKeys(t *testing.T) {
	dir := setupNormalizeVault(t)
	src := filepath.Join(dir, ".env")
	writeNormalizeEnv(t, src, "myKey=value\nanother_key=foo\n")

	dst := filepath.Join(dir, ".env.out")
	if err := Normalize(src, dst, true); err != nil {
		t.Fatalf("Normalize: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "MYKEY=") {
		t.Errorf("expected MYKEY uppercased, got: %s", out)
	}
	if !strings.Contains(out, "ANOTHER_KEY=") {
		t.Errorf("expected ANOTHER_KEY uppercased, got: %s", out)
	}
}
