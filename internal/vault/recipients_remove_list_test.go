package vault

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

func setupInfoVault(t *testing.T) (Config, func()) {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
		EncryptedFile:  filepath.Join(dir, ".env.age"),
		PlainFile:      filepath.Join(dir, ".env"),
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return cfg, func() { os.RemoveAll(dir) }
}

func TestListRecipientsInfo_Empty(t *testing.T) {
	cfg, cleanup := setupInfoVault(t)
	defer cleanup()

	infos, err := ListRecipientsInfo(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(infos) != 0 {
		t.Fatalf("expected 0 infos, got %d", len(infos))
	}
}

func TestListRecipientsInfo_ReturnsAll(t *testing.T) {
	cfg, cleanup := setupInfoVault(t)
	defer cleanup()

	const n = 3
	for i := 0; i < n; i++ {
		id, err := age.GenerateX25519Identity()
		if err != nil {
			t.Fatalf("generate identity: %v", err)
		}
		if err := AddRecipientToFile(cfg.RecipientsFile, id.Recipient().String()); err != nil {
			t.Fatalf("AddRecipient: %v", err)
		}
	}

	infos, err := ListRecipientsInfo(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(infos) != n {
		t.Fatalf("expected %d infos, got %d", n, len(infos))
	}
	for _, info := range infos {
		if info.PublicKey == "" {
			t.Error("PublicKey must not be empty")
		}
		if info.Label == "" {
			t.Error("Label must not be empty")
		}
	}
}

func TestListRecipientsInfo_NotInitialised(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		VaultDir:       filepath.Join(dir, ".envault"),
		RecipientsFile: filepath.Join(dir, ".envault", "recipients"),
	}
	_, err := ListRecipientsInfo(cfg)
	if err != ErrNotInitialised {
		t.Fatalf("expected ErrNotInitialised, got %v", err)
	}
}
