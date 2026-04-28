package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSampleVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:   filepath.Join(dir, ".envault"),
		PlainFile:  filepath.Join(dir, ".env"),
		CipherFile: filepath.Join(dir, ".env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeSampleEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestSample_DefaultDst(t *testing.T) {
	cfg, _ := setupSampleVault(t)
	writeSampleEnv(t, cfg.PlainFile, "A=1\nB=2\nC=3\n")

	opts := SampleOptions{}
	if err := Sample(cfg, opts); err != nil {
		t.Fatalf("Sample: %v", err)
	}

	data, err := os.ReadFile(cfg.PlainFile + ".sample")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if !strings.HasSuffix(l, "=") {
			t.Errorf("expected key= format, got %q", l)
		}
	}
}

func TestSample_LimitN(t *testing.T) {
	cfg, _ := setupSampleVault(t)
	writeSampleEnv(t, cfg.PlainFile, "A=1\nB=2\nC=3\nD=4\n")

	dst := cfg.PlainFile + ".out"
	opts := SampleOptions{Dst: dst, N: 2}
	if err := Sample(cfg, opts); err != nil {
		t.Fatalf("Sample: %v", err)
	}

	data, _ := os.ReadFile(dst)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestSample_KeysOnly(t *testing.T) {
	cfg, _ := setupSampleVault(t)
	writeSampleEnv(t, cfg.PlainFile, "FOO=bar\nBAZ=qux\n")

	dst := cfg.PlainFile + ".keys"
	opts := SampleOptions{Dst: dst, Keys: true}
	if err := Sample(cfg, opts); err != nil {
		t.Fatalf("Sample: %v", err)
	}

	data, _ := os.ReadFile(dst)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if lines[0] != "FOO" || lines[1] != "BAZ" {
		t.Errorf("unexpected keys output: %v", lines)
	}
}

func TestSample_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:   filepath.Join(dir, ".envault"),
		PlainFile:  filepath.Join(dir, ".env"),
		CipherFile: filepath.Join(dir, ".env.age"),
	}
	err := Sample(cfg, SampleOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
