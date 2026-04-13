package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSplitVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeSplitEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestSplit_ByPrefix(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	src := filepath.Join(dir, ".env")
	writeSplitEnv(t, src, "# config\nDB_HOST=localhost\nDB_PORT=5432\nREDIS_URL=redis://localhost\nAPP_NAME=myapp\n")

	dbDest := filepath.Join(dir, "db.env")
	redisDest := filepath.Join(dir, "redis.env")
	appDest := filepath.Join(dir, "app.env")

	opts := SplitOptions{
		Prefixes:  map[string]string{"DB_": dbDest, "REDIS_": redisDest},
		Remainder: appDest,
	}
	if err := Split(cfg, src, opts); err != nil {
		t.Fatalf("split: %v", err)
	}

	assertFileContains(t, dbDest, "DB_HOST=localhost")
	assertFileContains(t, dbDest, "DB_PORT=5432")
	assertFileContains(t, redisDest, "REDIS_URL=redis://localhost")
	assertFileContains(t, appDest, "APP_NAME=myapp")
}

func TestSplit_CommentsInAllFiles(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	src := filepath.Join(dir, ".env")
	writeSplitEnv(t, src, "# shared comment\nDB_HOST=localhost\nAPP_NAME=myapp\n")

	dbDest := filepath.Join(dir, "db.env")
	appDest := filepath.Join(dir, "app.env")

	opts := SplitOptions{
		Prefixes:  map[string]string{"DB_": dbDest},
		Remainder: appDest,
	}
	if err := Split(cfg, src, opts); err != nil {
		t.Fatalf("split: %v", err)
	}

	assertFileContains(t, dbDest, "# shared comment")
	assertFileContains(t, appDest, "# shared comment")
}

func TestSplit_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Split(cfg, filepath.Join(dir, ".env"), SplitOptions{Remainder: filepath.Join(dir, "out.env")})
	if err == nil || !strings.Contains(err.Error(), "not initialised") {
		t.Fatalf("expected not-initialised error, got %v", err)
	}
}

func TestSplit_NoPrefixesNoRemainder(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	src := filepath.Join(dir, ".env")
	writeSplitEnv(t, src, "DB_HOST=localhost\n")

	err := Split(cfg, src, SplitOptions{})
	if err == nil || !strings.Contains(err.Error(), "no prefixes") {
		t.Fatalf("expected error about no prefixes, got %v", err)
	}
}

func TestSplit_SourceNotFound(t *testing.T) {
	cfg, dir := setupSplitVault(t)
	err := Split(cfg, filepath.Join(dir, "missing.env"), SplitOptions{Remainder: filepath.Join(dir, "out.env")})
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("%s: expected to contain %q, got:\n%s", path, substr, data)
	}
}
