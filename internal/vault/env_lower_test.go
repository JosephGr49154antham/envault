package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupLowerVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".envault"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeLowerEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLower_AllKeys(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "DB_HOST=LOCALHOST\nDB_PORT=5432\nAPP_NAME=MyApp\n")
	dst := filepath.Join(dir, ".env.lower")

	if err := Lower(src, dst, nil, false); err != nil {
		t.Fatalf("Lower error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "DB_HOST=localhost") {
		t.Errorf("expected DB_HOST=localhost, got: %s", out)
	}
	if !strings.Contains(out, "APP_NAME=myapp") {
		t.Errorf("expected APP_NAME=myapp, got: %s", out)
	}
}

func TestLower_SelectedKeys(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "DB_HOST=LOCALHOST\nDB_PORT=5432\nAPP_NAME=MyApp\n")
	dst := filepath.Join(dir, ".env.lower")

	if err := Lower(src, dst, []string{"DB_HOST"}, false); err != nil {
		t.Fatalf("Lower error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "DB_HOST=localhost") {
		t.Errorf("expected DB_HOST=localhost")
	}
	if !strings.Contains(out, "APP_NAME=MyApp") {
		t.Errorf("expected APP_NAME unchanged")
	}
}

func TestLower_CustomDst(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "KEY=VALUE\n")
	dst := filepath.Join(dir, "custom.env")

	if err := Lower(src, dst, nil, false); err != nil {
		t.Fatalf("Lower error: %v", err)
	}

	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected custom dst to exist")
	}
}

func TestLower_PreservesCommentsAndBlanks(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "# comment\n\nKEY=VALUE\n")
	dst := filepath.Join(dir, ".env.lower")

	if err := Lower(src, dst, nil, false); err != nil {
		t.Fatalf("Lower error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !strings.Contains(out, "# comment") {
		t.Errorf("expected comment preserved")
	}
	if !strings.Contains(out, "\n\n") {
		t.Errorf("expected blank line preserved")
	}
}

func TestLower_InPlace(t *testing.T) {
	dir := setupLowerVault(t)
	src := writeLowerEnv(t, dir, ".env", "KEY=UPPER_VALUE\n")

	if err := Lower(src, src, nil, true); err != nil {
		t.Fatalf("Lower in-place error: %v", err)
	}

	data, _ := os.ReadFile(src)
	if !strings.Contains(string(data), "KEY=upper_value") {
		t.Errorf("expected in-place lower, got: %s", string(data))
	}
}
