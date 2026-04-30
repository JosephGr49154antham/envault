package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupQuoteVault(t *testing.T) (Config, string) {
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

func writeQuoteEnv(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
}

func TestQuote_QuotesUnquotedValues(t *testing.T) {
	cfg, dir := setupQuoteVault(t)
	src := filepath.Join(dir, ".env")
	writeQuoteEnv(t, src, "KEY=hello\nDB=world\n")

	if err := Quote(cfg, QuoteOptions{Src: src}); err != nil {
		t.Fatalf("Quote: %v", err)
	}
	got, _ := os.ReadFile(src)
	if string(got) != `KEY="hello"`+"\n"+`DB="world"`+"\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestQuote_PreservesAlreadyQuoted(t *testing.T) {
	cfg, dir := setupQuoteVault(t)
	src := filepath.Join(dir, ".env")
	writeQuoteEnv(t, src, `KEY="already"`+"\n")

	if err := Quote(cfg, QuoteOptions{Src: src}); err != nil {
		t.Fatalf("Quote: %v", err)
	}
	got, _ := os.ReadFile(src)
	if string(got) != `KEY="already"`+"\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestQuote_ForceRequotes(t *testing.T) {
	cfg, dir := setupQuoteVault(t)
	src := filepath.Join(dir, ".env")
	writeQuoteEnv(t, src, `KEY="old"`+"\n")

	if err := Quote(cfg, QuoteOptions{Src: src, Force: true}); err != nil {
		t.Fatalf("Quote: %v", err)
	}
	got, _ := os.ReadFile(src)
	if string(got) != `KEY="old"`+"\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestQuote_CustomDst(t *testing.T) {
	cfg, dir := setupQuoteVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.quoted")
	writeQuoteEnv(t, src, "FOO=bar\n")

	if err := Quote(cfg, QuoteOptions{Src: src, Dst: dst}); err != nil {
		t.Fatalf("Quote: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != `FOO="bar"`+"\n" {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestQuote_NoOverwrite(t *testing.T) {
	cfg, dir := setupQuoteVault(t)
	src := filepath.Join(dir, ".env")
	dst := filepath.Join(dir, ".env.out")
	writeQuoteEnv(t, src, "A=1\n")
	writeQuoteEnv(t, dst, "existing\n")

	err := Quote(cfg, QuoteOptions{Src: src, Dst: dst})
	if err == nil {
		t.Fatal("expected error for existing dst")
	}
}

func TestQuote_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	err := Quote(cfg, QuoteOptions{Src: filepath.Join(dir, ".env")})
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}
