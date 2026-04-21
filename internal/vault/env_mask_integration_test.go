package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMask_CommentsAndBlanksPreserved(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")

	content := "# database config\nDB_PASSWORD=secret\n\nDB_HOST=localhost\n"
	writeMaskEnv(t, src, content)

	if err := Mask(cfg, MaskOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("Mask: %v", err)
	}

	data, _ := os.ReadFile(dst)
	result := string(data)

	if !strings.Contains(result, "# database config") {
		t.Error("expected comment line to be preserved")
	}
	if !strings.Contains(result, "DB_HOST=localhost") {
		t.Error("expected non-sensitive value to be preserved")
	}
	if strings.Contains(result, "secret") {
		t.Error("expected DB_PASSWORD value to be masked")
	}
}

func TestMask_EmptyValue_LeftAlone(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")

	writeMaskEnv(t, src, "API_KEY=\n")

	if err := Mask(cfg, MaskOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("Mask: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "API_KEY=") {
		t.Error("expected empty value line to be present")
	}
}

func TestMask_OverwriteFlag(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")

	writeMaskEnv(t, src, "TOKEN=abc\n")
	_ = os.WriteFile(dst, []byte("old content"), 0o600)

	if err := Mask(cfg, MaskOptions{Src: src, Dst: dst, Overwrite: true}); err != nil {
		t.Fatalf("Mask with overwrite: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if strings.Contains(string(data), "old content") {
		t.Error("expected destination to be overwritten")
	}
	if strings.Contains(string(data), "abc") {
		t.Error("expected TOKEN value to be masked after overwrite")
	}
}
