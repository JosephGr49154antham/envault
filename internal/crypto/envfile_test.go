package crypto_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"

	"github.com/yourusername/envault/internal/crypto"
)

func TestEncryptDecryptEnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	id, recipient := generateTestRecipientIdentity(t)

	envContent := []byte("API_KEY=abc123\nDB_URL=postgres://localhost/mydb\n")
	srcPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(srcPath, envContent, 0o600); err != nil {
		t.Fatalf("writing test .env: %v", err)
	}

	encPath := filepath.Join(tmpDir, ".env.age")
	if err := crypto.EncryptEnvFile(srcPath, encPath, []age.Recipient{recipient}); err != nil {
		t.Fatalf("EncryptEnvFile: %v", err)
	}

	if _, err := os.Stat(encPath); err != nil {
		t.Fatalf("encrypted file not found: %v", err)
	}

	decPath := filepath.Join(tmpDir, ".env.decrypted")
	if err := crypto.DecryptEnvFile(encPath, decPath, []age.Identity{id}); err != nil {
		t.Fatalf("DecryptEnvFile: %v", err)
	}

	got, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("reading decrypted file: %v", err)
	}

	if string(got) != string(envContent) {
		t.Errorf("content mismatch: got %q, want %q", got, envContent)
	}
}

func TestEncryptEnvFile_DefaultDst(t *testing.T) {
	tmpDir := t.TempDir()
	_, recipient := generateTestRecipientIdentity(t)

	srcPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(srcPath, []byte("X=1"), 0o600); err != nil {
		t.Fatalf("writing test .env: %v", err)
	}

	if err := crypto.EncryptEnvFile(srcPath, "", []age.Recipient{recipient}); err != nil {
		t.Fatalf("EncryptEnvFile with default dst: %v", err)
	}

	expected := srcPath + ".age"
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected encrypted file at %q, got error: %v", expected, err)
	}
}

func TestEncryptEnvFile_NotFound(t *testing.T) {
	_, recipient := generateTestRecipientIdentity(t)
	err := crypto.EncryptEnvFile("/nonexistent/.env", "", []age.Recipient{recipient})
	if err == nil {
		t.Fatal("expected error for missing source file, got nil")
	}
}
