package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupConvertVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeConvertEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestConvert_ToDotenvExport(t *testing.T) {
	cfg, dir := setupConvertVault(t)
	src := filepath.Join(dir, ".env")
	writeConvertEnv(t, src, "FOO=bar\nBAZ=qux\n")

	dst, err := Convert(cfg, ConvertOptions{Src: src, Format: FormatExport})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "export FOO=") {
		t.Errorf("expected export prefix, got: %s", string(data))
	}
}

func TestConvert_ToJSON(t *testing.T) {
	cfg, dir := setupConvertVault(t)
	src := filepath.Join(dir, ".env")
	writeConvertEnv(t, src, "API_KEY=secret\nDEBUG=true\n")

	dst, err := Convert(cfg, ConvertOptions{Src: src, Format: FormatJSON})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(dst, ".json") {
		t.Errorf("expected .json extension, got %s", dst)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "{") {
		t.Errorf("expected JSON braces, got: %s", string(data))
	}
}

func TestConvert_ToYAML(t *testing.T) {
	cfg, dir := setupConvertVault(t)
	src := filepath.Join(dir, ".env")
	writeConvertEnv(t, src, "HOST=localhost\nPORT=5432\n")

	dst, err := Convert(cfg, ConvertOptions{Src: src, Format: FormatYAML})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "HOST:") {
		t.Errorf("expected YAML key, got: %s", string(data))
	}
}

func TestConvert_CustomDst(t *testing.T) {
	cfg, dir := setupConvertVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, "out.json")
	writeConvertEnv(t, src, "X=1\n")

	got, err := Convert(cfg, ConvertOptions{Src: src, Dst: dst, Format: FormatJSON})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dst {
		t.Errorf("expected %s, got %s", dst, got)
	}
}

func TestConvert_NoOverwrite(t *testing.T) {
	cfg, dir := setupConvertVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, "out.json")
	writeConvertEnv(t, src, "X=1\n")
	_ = os.WriteFile(dst, []byte("{}"), 0o600)

	_, err := Convert(cfg, ConvertOptions{Src: src, Dst: dst, Format: FormatJSON})
	if err == nil {
		t.Error("expected error for existing destination")
	}
}

func TestConvert_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Convert(cfg, ConvertOptions{Src: filepath.Join(dir, ".env"), Format: FormatJSON})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}
