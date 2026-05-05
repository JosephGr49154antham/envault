package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCommentVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "secrets.env.age"),
	}
	_ = os.MkdirAll(cfg.VaultDir, 0o700)
	return cfg, dir
}

func writeCommentEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestComment_CommentsOutAllKeys(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	writeCommentEnv(t, src, "FOO=bar\nBAZ=qux\n")

	dst := filepath.Join(dir, "out.env")
	err := Comment(cfg, CommentOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := os.ReadFile(dst)
	if !strings.Contains(string(b), "# FOO=bar") {
		t.Errorf("expected FOO to be commented, got:\n%s", b)
	}
	if !strings.Contains(string(b), "# BAZ=qux") {
		t.Errorf("expected BAZ to be commented, got:\n%s", b)
	}
}

func TestComment_SelectedKeys(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	writeCommentEnv(t, src, "FOO=bar\nBAZ=qux\n")

	dst := filepath.Join(dir, "out.env")
	err := Comment(cfg, CommentOptions{Src: src, Dst: dst, Keys: []string{"FOO"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := os.ReadFile(dst)
	if !strings.Contains(string(b), "# FOO=bar") {
		t.Errorf("expected FOO commented")
	}
	if strings.Contains(string(b), "# BAZ") {
		t.Errorf("BAZ should not be commented")
	}
}

func TestComment_Uncomment(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	writeCommentEnv(t, src, "# FOO=bar\nBAZ=qux\n")

	dst := filepath.Join(dir, "out.env")
	err := Comment(cfg, CommentOptions{Src: src, Dst: dst, Uncomment: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := os.ReadFile(dst)
	if !strings.Contains(string(b), "FOO=bar") || strings.Contains(string(b), "# FOO") {
		t.Errorf("expected FOO to be uncommented, got:\n%s", b)
	}
}

func TestComment_NoOverwrite(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	writeCommentEnv(t, src, "FOO=bar\n")
	dst := filepath.Join(dir, "out.env")
	writeCommentEnv(t, dst, "existing\n")

	err := Comment(cfg, CommentOptions{Src: src, Dst: dst})
	if err == nil {
		t.Fatal("expected error for existing dst")
	}
}

func TestComment_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{VaultDir: filepath.Join(dir, ".envault")}
	err := Comment(cfg, CommentOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestComment_DefaultDst(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, "secrets.env")
	writeCommentEnv(t, src, "KEY=val\n")

	err := Comment(cfg, CommentOptions{Src: src})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(dir, "secrets.commented.env")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected default dst %s to exist", expected)
	}
}
