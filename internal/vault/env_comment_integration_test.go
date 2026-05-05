package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComment_PreservesBlankLinesAndComments(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	input := "# header\n\nFOO=bar\n# standalone comment\nBAZ=qux\n"
	writeCommentEnv(t, src, input)

	dst := filepath.Join(dir, "out.env")
	if err := Comment(cfg, CommentOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, _ := os.ReadFile(dst)
	got := string(b)

	if !strings.Contains(got, "# header") {
		t.Error("header comment should be preserved")
	}
	if !strings.Contains(got, "# standalone comment") {
		t.Error("standalone comment should be preserved")
	}
	if !strings.Contains(got, "# FOO=bar") {
		t.Error("FOO should be commented out")
	}
	if !strings.Contains(got, "# BAZ=qux") {
		t.Error("BAZ should be commented out")
	}
}

func TestComment_RoundTrip(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	original := "ALPHA=one\nBETA=two\n"
	writeCommentEnv(t, src, original)

	commented := filepath.Join(dir, "commented.env")
	if err := Comment(cfg, CommentOptions{Src: src, Dst: commented}); err != nil {
		t.Fatalf("comment: %v", err)
	}

	restored := filepath.Join(dir, "restored.env")
	if err := Comment(cfg, CommentOptions{Src: commented, Dst: restored, Uncomment: true}); err != nil {
		t.Fatalf("uncomment: %v", err)
	}

	b, _ := os.ReadFile(restored)
	got := string(b)
	if !strings.Contains(got, "ALPHA=one") || !strings.Contains(got, "BETA=two") {
		t.Errorf("round-trip failed, got:\n%s", got)
	}
}

func TestComment_OverwriteFlag(t *testing.T) {
	cfg, dir := setupCommentVault(t)
	src := filepath.Join(dir, ".env")
	writeCommentEnv(t, src, "SECRET=abc\n")

	// First pass: create dst
	dst := filepath.Join(dir, "out.env")
	if err := Comment(cfg, CommentOptions{Src: src, Dst: dst, Overwrite: true}); err != nil {
		t.Fatalf("first pass: %v", err)
	}
	// Second pass: overwrite should succeed
	if err := Comment(cfg, CommentOptions{Src: src, Dst: dst, Overwrite: true}); err != nil {
		t.Fatalf("overwrite pass: %v", err)
	}
	b, _ := os.ReadFile(dst)
	if !strings.Contains(string(b), "# SECRET=abc") {
		t.Errorf("expected commented output, got: %s", b)
	}
}
