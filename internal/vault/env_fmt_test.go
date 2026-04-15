package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupFmtVault(t *testing.T) (Config, string) {
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

func writeFmtEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	return p
}

func TestFmt_TrimsValues(t *testing.T) {
	cfg, dir := setupFmtVault(t)
	src := writeFmtEnv(t, dir, ".env", "KEY=  hello  \nOTHER= world\n")

	if err := Fmt(cfg, src, FmtOptions{TrimValues: true}); err != nil {
		t.Fatalf("Fmt: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != "KEY=hello\nOTHER=world\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestFmt_QuotesValuesWithSpaces(t *testing.T) {
	cfg, dir := setupFmtVault(t)
	src := writeFmtEnv(t, dir, ".env", "MSG world\n")

	if err := Fmt(cfg, src, FmtOptions{TrimValues: true, QuoteValues: true}); err != nil {
		t.Fatalf("Fmt: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != `MSG="hello world"`+"\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestFmt_PreservesCommentsAndBlanks(t *testing.T) {
	cfg, dir := setupFmtVault(t)
	input := "# header\n\nKEY=value\n"
	src := writeFmtEnv(t, dir, ".env", input)

	if err := Fmt(cfg, src, FmtOptions{}); err != nil {
		t.Fatalf("Fmt: %v", err)
	}

	got, _ := os.ReadFile(src)
	if string(got) != input {
		t.Errorf("expected unchanged, got:\n%s", got)
	}
}

func TestFmt_CustomDst(t *testing.T) {
	cfg, dir := setupFmtVault(t)
	src := writeFmtEnv(t, dir, ".env", "KEY=  val  \n")
	dst := filepath.Join(dir, ".env.fmt")

	if err := Fmt(cfg, src, FmtOptions{Dst: dst, TrimValues: true}); err != nil {
		t.Fatalf("Fmt: %v", err)
	}

	got, _ := os.ReadFile(dst)
	if string(got) != "KEY=val\n" {
		t.Errorf("unexpected dst content:\n%s", got)
	}
	// source should be unchanged
	orig, _ := os.ReadFile(src)
	if string(orig) != "KEY=  val  \n" {
		t.Errorf("source was modified unexpectedly")
	}
}

func TestFmt_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Fmt(cfg, filepath.Join(dir, ".env"), FmtOptions{})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
