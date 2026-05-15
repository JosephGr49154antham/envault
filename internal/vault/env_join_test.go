package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupJoinVault(t *testing.T) (Config, string) {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(tmpDir, ".envault"),
		RecipientsFile: filepath.Join(tmpDir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(tmpDir, ".envault", "secrets.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, tmpDir
}

func writeJoinEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestJoin_MergesFiles(t *testing.T) {
	cfg, dir := setupJoinVault(t)
	a := filepath.Join(dir, ".env.a")
	b := filepath.Join(dir, ".env.b")
	writeJoinEnv(t, a, "FOO=1\nBAR=2\n")
	writeJoinEnv(t, b, "BAZ=3\nQUX=4\n")

	dst := filepath.Join(dir, "joined.env")
	err := Join(cfg, []string{a, b}, JoinOptions{Dst: dst})
	if err != nil {
		t.Fatalf("Join: %v", err)
	}

	data, _ := os.ReadFile(dst)
	content := string(data)
	for _, key := range []string{"FOO=1", "BAR=2", "BAZ=3", "QUX=4"} {
		if !strings.Contains(content, key) {
			t.Errorf("expected %q in output", key)
		}
	}
}

func TestJoin_DefaultDst(t *testing.T) {
	cfg, dir := setupJoinVault(t)
	a := filepath.Join(dir, ".env.a")
	writeJoinEnv(t, a, "KEY=val\n")

	err := Join(cfg, []string{a}, JoinOptions{})
	if err != nil {
		t.Fatalf("Join: %v", err)
	}

	defaultDst := filepath.Join(cfg.VaultDir, ".env.joined")
	if _, err := os.Stat(defaultDst); os.IsNotExist(err) {
		t.Errorf("expected default dst %s to exist", defaultDst)
	}
}

func TestJoin_NoOverwrite(t *testing.T) {
	cfg, dir := setupJoinVault(t)
	a := filepath.Join(dir, ".env.a")
	writeJoinEnv(t, a, "A=1\n")
	dst := filepath.Join(dir, "out.env")
	writeJoinEnv(t, dst, "existing\n")

	err := Join(cfg, []string{a}, JoinOptions{Dst: dst, Overwrite: false})
	if err == nil {
		t.Error("expected error when dst exists and overwrite=false")
	}
}

func TestJoin_WithSeparator(t *testing.T) {
	cfg, dir := setupJoinVault(t)
	a := filepath.Join(dir, ".env.a")
	b := filepath.Join(dir, ".env.b")
	writeJoinEnv(t, a, "A=1\n")
	writeJoinEnv(t, b, "B=2\n")
	dst := filepath.Join(dir, "sep.env")

	err := Join(cfg, []string{a, b}, JoinOptions{Dst: dst, Separator: true})
	if err != nil {
		t.Fatalf("Join: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "# --- .env.a ---") {
		t.Error("expected separator comment for .env.a")
	}
}

func TestJoin_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/nope"}
	err := Join(cfg, []string{".env"}, JoinOptions{})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestJoin_NoSources(t *testing.T) {
	cfg, _ := setupJoinVault(t)
	err := Join(cfg, nil, JoinOptions{})
	if err == nil {
		t.Error("expected error with no source files")
	}
}
