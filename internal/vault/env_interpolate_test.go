package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupInterpolateVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}
	return cfg, dir
}

func writeInterpolateEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestInterpolate_ExpandsLocalReference(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	src := filepath.Join(dir, ".env")
	writeInterpolateEnv(t, src, "BASE=/opt/app\nPATH=${BASE}/bin\n")

	opts := InterpolateOptions{Src: src}
	if err := Interpolate(cfg, opts); err != nil {
		t.Fatalf("Interpolate: %v", err)
	}

	dst := strings.TrimSuffix(src, ".env") + ".interpolated.env"
	out, _ := os.ReadFile(dst)
	if !strings.Contains(string(out), "PATH=/opt/app/bin") {
		t.Errorf("expected expanded PATH, got:\n%s", out)
	}
}

func TestInterpolate_OverlayTakesPrecedence(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	src := filepath.Join(dir, ".env")
	overlay := filepath.Join(dir, ".env.local")
	writeInterpolateEnv(t, src, "HOST=localhost\nURL=http://${HOST}/api\n")
	writeInterpolateEnv(t, overlay, "HOST=prod.example.com\n")

	dst := filepath.Join(dir, "out.env")
	opts := InterpolateOptions{Src: src, Overlay: overlay, Dst: dst}
	if err := Interpolate(cfg, opts); err != nil {
		t.Fatalf("Interpolate: %v", err)
	}

	out, _ := os.ReadFile(dst)
	if !strings.Contains(string(out), "URL=http://prod.example.com/api") {
		t.Errorf("expected overlay HOST, got:\n%s", out)
	}
}

func TestInterpolate_StrictMode_UndefinedVar(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	src := filepath.Join(dir, ".env")
	writeInterpolateEnv(t, src, "URL=http://${UNDEFINED_XYZ_123}/path\n")

	// ensure env var is not set
	os.Unsetenv("UNDEFINED_XYZ_123")

	opts := InterpolateOptions{Src: src, Strict: true}
	if err := Interpolate(cfg, opts); err == nil {
		t.Error("expected error in strict mode for undefined variable")
	}
}

func TestInterpolate_CommentsPreserved(t *testing.T) {
	cfg, dir := setupInterpolateVault(t)
	src := filepath.Join(dir, ".env")
	writeInterpolateEnv(t, src, "# header comment\nFOO=bar\n# another comment\n")

	dst := filepath.Join(dir, "result.env")
	opts := InterpolateOptions{Src: src, Dst: dst}
	if err := Interpolate(cfg, opts); err != nil {
		t.Fatalf("Interpolate: %v", err)
	}

	out, _ := os.ReadFile(dst)
	if !strings.Contains(string(out), "# header comment") {
		t.Errorf("expected comments preserved, got:\n%s", out)
	}
}

func TestInterpolate_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:    filepath.Join(dir, ".envault"),
		PlainFile:   filepath.Join(dir, ".env"),
		EncryptedFile: filepath.Join(dir, ".env.age"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	err := Interpolate(cfg, InterpolateOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Error("expected error for uninitialised vault")
	}
}
