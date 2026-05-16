package vault_test

import (
	"os"
	"path/filepath"
	"testing"
)

func setupUnwrapVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := Init(DefaultConfig(dir)); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir
}

func writeUnwrapEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeUnwrapEnv: %v", err)
	}
}

func TestUnwrap_JoinsContinuationLines(t *testing.T) {
	dir := setupUnwrapVault(t)
	src := filepath.Join(dir, ".env")
	writeUnwrapEnv(t, src,
		"DB_URL=postgres://localhost/\\
			mydb
			API_KEY=abc123
			")

	dst := filepath.Join(dir, ".env.unwrapped")
	if err := Unwrap(dir, src, dst, false); err != nil {
		t.Fatalf("Unwrap: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	out := string(data)
	if !contains(out, "DB_URL=postgres://localhost/mydb") {
		t.Errorf("expected joined DB_URL line, got:\n%s", out)
	}
	if !contains(out, "API_KEY=abc123") {
		t.Errorf("expected API_KEY line, got:\n%s", out)
	}
}

func TestUnwrap_DefaultDst(t *testing.T) {
	dir := setupUnwrapVault(t)
	src := filepath.Join(dir, ".env.wrapped")
	writeUnwrapEnv(t, src, "KEY=value
")

	if err := Unwrap(dir, src, "", false); err != nil {
		t.Fatalf("Unwrap: %v", err)
	}

	expected := filepath.Join(dir, ".env")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected default dst %s to exist", expected)
	}
}

func TestUnwrap_NoOverwrite(t *testing.T) {
	dir := setupUnwrapVault(t)
	src := filepath.Join(dir, ".env.wrapped")
	writeUnwrapEnv(t, src, "KEY=value
")

	dst := filepath.Join(dir, ".env.out")
	writeUnwrapEnv(t, dst, "existing=content
")

	err := Unwrap(dir, src, dst, false)
	if err == nil {
		t.Error("expected error when dst exists and overwrite=false")
	}
}

func TestUnwrap_OverwriteFlag(t *testing.T) {
	dir := setupUnwrapVault(t)
	src := filepath.Join(dir, ".env.wrapped")
	writeUnwrapEnv(t, src, "KEY=newvalue
")

	dst := filepath.Join(dir, ".env.out")
	writeUnwrapEnv(t, dst, "KEY=oldvalue
")

	if err := Unwrap(dir, src, dst, true); err != nil {
		t.Fatalf("Unwrap with overwrite: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !contains(string(data), "newvalue") {
		t.Errorf("expected overwritten value, got: %s", data)
	}
}

func TestUnwrap_PreservesCommentsAndBlanks(t *testing.T) {
	dir := setupUnwrapVault(t)
	src := filepath.Join(dir, ".env")
	writeUnwrapEnv(t, src,
		"# database config

DB_HOST=localhost
DB_PORT=5432
")

	dst := filepath.Join(dir, ".env.out")
	if err := Unwrap(dir, src, dst, false); err != nil {
		t.Fatalf("Unwrap: %v", err)
	}

	data, _ := os.ReadFile(dst)
	out := string(data)
	if !contains(out, "# database config") {
		t.Errorf("expected comment to be preserved, got:\n%s", out)
	}
	if !contains(out, "DB_HOST=localhost") {
		t.Errorf("expected DB_HOST line, got:\n%s", out)
	}
}

func TestUnwrap_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	err := Unwrap(dir, filepath.Join(dir, ".env"), "", false)
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}
