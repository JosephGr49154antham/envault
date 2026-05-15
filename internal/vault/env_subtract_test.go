package vault_test

import (
	"os"
	"path/filepath"
	"testing"
)

func setupSubtractVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := Init(DefaultConfig(dir)); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir
}

func writeSubtractEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestSubtract_RemovesMatchingKeys(t *testing.T) {
	dir := setupSubtractVault(t)

	src := filepath.Join(dir, ".env")
	writeSubtractEnv(t, src, "APP_NAME=myapp\nDEBUG=true\nSECRET=abc123\nPORT=8080\n")

	exclude := filepath.Join(dir, ".env.exclude")
	writeSubtractEnv(t, exclude, "DEBUG=anything\nSECRET=anything\n")

	dst := filepath.Join(dir, ".env.out")

	err := Subtract(SubtractOptions{
		Root:    dir,
		Src:     src,
		Exclude: exclude,
		Dst:     dst,
	})
	if err != nil {
		t.Fatalf("Subtract: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}

	got := string(data)
	if contains(got, "DEBUG") {
		t.Error("expected DEBUG to be removed")
	}
	if contains(got, "SECRET") {
		t.Error("expected SECRET to be removed")
	}
	if !contains(got, "APP_NAME") {
		t.Error("expected APP_NAME to be retained")
	}
	if !contains(got, "PORT") {
		t.Error("expected PORT to be retained")
	}
}

func TestSubtract_NoOverwrite(t *testing.T) {
	dir := setupSubtractVault(t)

	src := filepath.Join(dir, ".env")
	writeSubtractEnv(t, src, "KEY=value\n")

	exclude := filepath.Join(dir, ".env.exclude")
	writeSubtractEnv(t, exclude, "OTHER=x\n")

	dst := filepath.Join(dir, ".env.out")
	writeSubtractEnv(t, dst, "existing=content\n")

	err := Subtract(SubtractOptions{
		Root:      dir,
		Src:       src,
		Exclude:   exclude,
		Dst:       dst,
		Overwrite: false,
	})
	if err == nil {
		t.Fatal("expected error when dst exists and overwrite=false")
	}
}

func TestSubtract_NotInitialised(t *testing.T) {
	dir := t.TempDir()

	err := Subtract(SubtractOptions{
		Root:    dir,
		Src:     filepath.Join(dir, ".env"),
		Exclude: filepath.Join(dir, ".env.exclude"),
	})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestSubtract_DefaultDst(t *testing.T) {
	dir := setupSubtractVault(t)

	src := filepath.Join(dir, ".env")
	writeSubtractEnv(t, src, "A=1\nB=2\nC=3\n")

	exclude := filepath.Join(dir, ".env.exclude")
	writeSubtractEnv(t, exclude, "B=ignored\n")

	err := Subtract(SubtractOptions{
		Root:    dir,
		Src:     src,
		Exclude: exclude,
	})
	if err != nil {
		t.Fatalf("Subtract: %v", err)
	}

	defaultDst := filepath.Join(dir, ".env.subtracted")
	if _, err := os.Stat(defaultDst); os.IsNotExist(err) {
		t.Errorf("expected default dst %s to be created", defaultDst)
	}
}

func TestSubtract_PreservesComments(t *testing.T) {
	dir := setupSubtractVault(t)

	src := filepath.Join(dir, ".env")
	writeSubtractEnv(t, src, "# app config\nAPP=myapp\n# debug flag\nDEBUG=true\nPORT=9000\n")

	exclude := filepath.Join(dir, ".env.exclude")
	writeSubtractEnv(t, exclude, "DEBUG=x\n")

	dst := filepath.Join(dir, ".env.out")

	err := Subtract(SubtractOptions{
		Root:    dir,
		Src:     src,
		Exclude: exclude,
		Dst:     dst,
	})
	if err != nil {
		t.Fatalf("Subtract: %v", err)
	}

	data, _ := os.ReadFile(dst)
	got := string(data)
	if !contains(got, "# app config") {
		t.Error("expected header comment to be preserved")
	}
	if !contains(got, "APP=myapp") {
		t.Error("expected APP to be retained")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
