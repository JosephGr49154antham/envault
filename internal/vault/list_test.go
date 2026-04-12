package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestList_EmptyVault(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:     filepath.Join(dir, ".envault"),
		EncryptedDir: filepath.Join(dir, ".envault", "encrypted"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}

	infos, err := List(cfg)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 entries, got %d", len(infos))
	}
}

func TestList_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedDir:   filepath.Join(dir, ".envault", "encrypted"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	_, err := List(cfg)
	if err == nil {
		t.Fatal("expected error for uninitialised vault, got nil")
	}
}

func TestList_ReturnsAgeFiles(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedDir:   filepath.Join(dir, ".envault", "encrypted"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Write a fake .age file and a non-.age file that should be ignored.
	if err := os.WriteFile(filepath.Join(cfg.EncryptedDir, ".env.age"), []byte("data"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.EncryptedDir, "README.txt"), []byte("ignore"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	infos, err := List(cfg)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(infos))
	}
	if infos[0].Name != ".env" {
		t.Errorf("expected name '.env', got %q", infos[0].Name)
	}
	if !infos[0].HasEnc {
		t.Error("expected HasEnc to be true")
	}
	if infos[0].EncMtime.IsZero() {
		t.Error("expected non-zero EncMtime")
	}
	_ = time.Now() // ensure time package used
}

func TestList_DetectsPlainFile(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		EncryptedDir:   filepath.Join(dir, ".envault", "encrypted"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients.txt"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}

	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("KEY=val"), 0600); err != nil {
		t.Fatalf("WriteFile plain: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.EncryptedDir, ".env.age"), []byte("enc"), 0600); err != nil {
		t.Fatalf("WriteFile enc: %v", err)
	}

	// Change to dir so relative plain path resolves.
	old, _ := os.Getwd()
	defer os.Chdir(old) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	infos, err := List(cfg)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(infos))
	}
	if !infos[0].HasPlain {
		t.Error("expected HasPlain to be true when .env exists")
	}
}
