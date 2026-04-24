package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupResolveVault(t *testing.T) (Config, string) {
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
	return cfg, dir
}

func writeResolveEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestResolve_ExpandsLocalReference(t *testing.T) {
	cfg, _ := setupResolveVault(t)
	writeResolveEnv(t, cfg.PlainFile, "BASE=/opt/app\nBIN=${BASE}/bin\n")

	if err := Resolve(cfg, ResolveOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(cfg.PlainFile)
	if !strings.Contains(string(data), "BIN=/opt/app/bin") {
		t.Errorf("expected expanded value, got:\n%s", data)
	}
}

func TestResolve_ExpandsEnvVar(t *testing.T) {
	cfg, _ := setupResolveVault(t)
	t.Setenv("INJECTED", "hello")
	writeResolveEnv(t, cfg.PlainFile, "GREETING=${INJECTED}\n")

	if err := Resolve(cfg, ResolveOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(cfg.PlainFile)
	if !strings.Contains(string(data), "GREETING=hello") {
		t.Errorf("expected GREETING=hello, got:\n%s", data)
	}
}

func TestResolve_LeavesUnknownIntact(t *testing.T) {
	cfg, _ := setupResolveVault(t)
	writeResolveEnv(t, cfg.PlainFile, "FOO=${UNKNOWN_XYZ}\n")

	if err := Resolve(cfg, ResolveOptions{Strict: false}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(cfg.PlainFile)
	if !strings.Contains(string(data), "${UNKNOWN_XYZ}") {
		t.Errorf("expected reference left intact, got:\n%s", data)
	}
}

func TestResolve_StrictMode_ReturnsError(t *testing.T) {
	cfg, _ := setupResolveVault(t)
	writeResolveEnv(t, cfg.PlainFile, "FOO=${DEFINITELY_NOT_SET_XYZ}\n")

	err := Resolve(cfg, ResolveOptions{Strict: true})
	if err == nil {
		t.Fatal("expected error in strict mode, got nil")
	}
	if !strings.Contains(err.Error(), "unresolvable") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestResolve_CustomSrcDst(t *testing.T) {
	cfg, dir := setupResolveVault(t)
	src := filepath.Join(dir, "input.env")
	dst := filepath.Join(dir, "output.env")
	writeResolveEnv(t, src, "ROOT=/srv\nDATA=${ROOT}/data\n")

	if err := Resolve(cfg, ResolveOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "DATA=/srv/data") {
		t.Errorf("expected DATA=/srv/data, got:\n%s", data)
	}
}

func TestResolve_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	err := Resolve(cfg, ResolveOptions{})
	if err == nil || !strings.Contains(err.Error(), "not initialised") {
		t.Errorf("expected not-initialised error, got: %v", err)
	}
}

func TestResolve_PreservesCommentsAndBlanks(t *testing.T) {
	cfg, _ := setupResolveVault(t)
	writeResolveEnv(t, cfg.PlainFile, "# header\n\nHOST=localhost\nURL=http://${HOST}/api\n")

	if err := Resolve(cfg, ResolveOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(cfg.PlainFile)
	if !strings.Contains(string(data), "# header") {
		t.Errorf("comment stripped, got:\n%s", data)
	}
	if !strings.Contains(string(data), "URL=http://localhost/api") {
		t.Errorf("expected expanded URL, got:\n%s", data)
	}
}
