package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTokenizeVault(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, dir
}

func writeTokenizeEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestTokenize_BasicParsing(t *testing.T) {
	cfg, dir := setupTokenizeVault(t)
	src := writeTokenizeEnv(t, dir, ".env", "# header\nFOO=bar\nBAZ=qux\n\n")

	res, err := Tokenize(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(res.Tokens))
	}
	if res.Tokens[0].Kind != "comment" {
		t.Errorf("expected comment, got %s", res.Tokens[0].Kind)
	}
	if res.Tokens[1].Key != "FOO" || res.Tokens[1].Value != "bar" {
		t.Errorf("unexpected token: %+v", res.Tokens[1])
	}
	if res.Tokens[3].Kind != "blank" {
		t.Errorf("expected blank, got %s", res.Tokens[3].Kind)
	}
}

func TestTokenize_InvalidLines(t *testing.T) {
	cfg, dir := setupTokenizeVault(t)
	src := writeTokenizeEnv(t, dir, ".env", "NOEQUALSSIGN\n=EMPTYKEY\n")

	res, err := Tokenize(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(res.Errors), res.Errors)
	}
	for _, tok := range res.Tokens {
		if tok.Kind != "invalid" {
			t.Errorf("expected invalid, got %s", tok.Kind)
		}
	}
}

func TestTokenize_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:      filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(dir, ".envault", "env.age"),
	}
	_, err := Tokenize(cfg, filepath.Join(dir, ".env"))
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestTokenize_QuotedValues(t *testing.T) {
	cfg, dir := setupTokenizeVault(t)
	src := writeTokenizeEnv(t, dir, ".env", `KEY="hello world"`+"\n")

	res, err := Tokenize(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(res.Tokens))
	}
	if res.Tokens[0].Value != "hello world" {
		t.Errorf("expected unquoted value, got %q", res.Tokens[0].Value)
	}
}
