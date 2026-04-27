package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupEchoVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncFile:     filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		IdentityFile:   filepath.Join(dir, ".envault", "identity.age"),
	}
	if err := os.MkdirAll(cfg.VaultDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return cfg, dir
}

func writeEchoEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestEcho_PrintsAllKeys(t *testing.T) {
	cfg, _ := setupEchoVault(t)
	writeEchoEnv(t, cfg.PlainFile, "FOO=bar\nBAZ=qux\n")

	// Capture stdout via pipe
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Echo(cfg, EchoOptions{})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sb strings.Builder
	buf := make([]byte, 256)
	for {
		n, e := r.Read(buf)
		sb.Write(buf[:n])
		if e != nil {
			break
		}
	}
	out := sb.String()
	if !strings.Contains(out, "FOO=bar") {
		t.Errorf("expected FOO=bar in output, got: %q", out)
	}
	if !strings.Contains(out, "BAZ=qux") {
		t.Errorf("expected BAZ=qux in output, got: %q", out)
	}
}

func TestEcho_SelectedKeys(t *testing.T) {
	cfg, _ := setupEchoVault(t)
	writeEchoEnv(t, cfg.PlainFile, "FOO=bar\nBAZ=qux\nSECRET=hidden\n")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Echo(cfg, EchoOptions{Keys: []string{"FOO"}})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sb strings.Builder
	buf := make([]byte, 256)
	for {
		n, e := r.Read(buf)
		sb.Write(buf[:n])
		if e != nil {
			break
		}
	}
	out := sb.String()
	if !strings.Contains(out, "FOO=bar") {
		t.Errorf("expected FOO=bar, got: %q", out)
	}
	if strings.Contains(out, "SECRET") {
		t.Errorf("SECRET should not appear in output, got: %q", out)
	}
}

func TestEcho_ExportFlag(t *testing.T) {
	cfg, _ := setupEchoVault(t)
	writeEchoEnv(t, cfg.PlainFile, "FOO=bar\n")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Echo(cfg, EchoOptions{Export: true})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sb strings.Builder
	buf := make([]byte, 256)
	for {
		n, e := r.Read(buf)
		sb.Write(buf[:n])
		if e != nil {
			break
		}
	}
	out := sb.String()
	if !strings.HasPrefix(out, "export FOO=bar") {
		t.Errorf("expected export prefix, got: %q", out)
	}
}

func TestEcho_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:  filepath.Join(dir, ".envault"),
		PlainFile: filepath.Join(dir, ".env"),
	}
	err := Echo(cfg, EchoOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
