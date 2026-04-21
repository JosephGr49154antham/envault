package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupMaskVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir: filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeMaskEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestMask_SensitiveValuesAreMasked(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")
	writeMaskEnv(t, src, "API_KEY=supersecret\nAPP_NAME=myapp\n")

	err := Mask(cfg, MaskOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("Mask: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if strings.Contains(string(data), "supersecret") {
		t.Error("expected value to be masked")
	}
	if !strings.Contains(string(data), "APP_NAME=myapp") {
		t.Error("expected non-sensitive value to be preserved")
	}
}

func TestMask_ExplicitKeys(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")
	writeMaskEnv(t, src, "APP_NAME=myapp\nPORT=8080\n")

	err := Mask(cfg, MaskOptions{Src: src, Dst: dst, Keys: []string{"PORT"}})
	if err != nil {
		t.Fatalf("Mask: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if strings.Contains(string(data), "8080") {
		t.Error("expected PORT value to be masked")
	}
	if !strings.Contains(string(data), "APP_NAME=myapp") {
		t.Error("expected APP_NAME to be preserved")
	}
}

func TestMask_DefaultDst(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	writeMaskEnv(t, src, "SECRET_KEY=abc123\n")

	err := Mask(cfg, MaskOptions{Src: src})
	if err != nil {
		t.Fatalf("Mask: %v", err)
	}

	expected := src + ".masked"
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected default dst %q to exist", expected)
	}
}

func TestMask_NoOverwrite(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")
	writeMaskEnv(t, src, "SECRET=val\n")
	_ = os.WriteFile(dst, []byte("existing"), 0o600)

	err := Mask(cfg, MaskOptions{Src: src, Dst: dst})
	if err == nil {
		t.Error("expected error when dst exists without --overwrite")
	}
}

func TestMask_NotInitialised(t *testing.T) {
	cfg := Config{VaultDir: t.TempDir() + "/missing"}
	err := Mask(cfg, MaskOptions{Src: ".env"})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}

func TestMask_CustomChar(t *testing.T) {
	cfg, dir := setupMaskVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.masked")
	writeMaskEnv(t, src, "DB_PASSWORD=hunter2\n")

	err := Mask(cfg, MaskOptions{Src: src, Dst: dst, MaskChar: "#"})
	if err != nil {
		t.Fatalf("Mask: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if !strings.Contains(string(data), "########") {
		t.Error("expected custom mask character '#'")
	}
}
