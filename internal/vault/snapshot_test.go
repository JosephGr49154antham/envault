package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/recipients"
	"github.com/nicholasgasior/envault/internal/vault"
)

func setupSnapshotVault(t *testing.T) (vault.Config, string) {
	t.Helper()
	tmp := t.TempDir()
	cfg := vault.Config{
		VaultDir:      filepath.Join(tmp, ".envault"),
		RecipientsFile: filepath.Join(tmp, ".envault", "recipients.txt"),
		EncryptedFile: filepath.Join(tmp, ".env.age"),
		IdentityFile:  filepath.Join(tmp, "identity.age"),
	}
	if err := vault.Init(cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	if err := keymgr.SaveIdentity(id, cfg.IdentityFile); err != nil {
		t.Fatalf("save identity: %v", err)
	}
	if err := recipients.AddRecipient(cfg.RecipientsFile, id.Recipient().String()); err != nil {
		t.Fatalf("add recipient: %v", err)
	}

	plain := "KEY1=hello\nKEY2=world\n"
	if err := crypto.EncryptEnvFile(cfg.EncryptedFile, []byte(plain), []string{id.Recipient().String()}); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return cfg, tmp
}

func TestSaveSnapshot_CreatesFile(t *testing.T) {
	cfg, _ := setupSnapshotVault(t)

	if err := vault.SaveSnapshot(cfg, "before-deploy"); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	snaps, err := vault.ListSnapshots(cfg)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snaps))
	}
	if snaps[0].Label != "before-deploy" {
		t.Errorf("label mismatch: got %q", snaps[0].Label)
	}
	if _, ok := snaps[0].Keys["KEY1"]; !ok {
		t.Error("expected KEY1 in snapshot")
	}
}

func TestListSnapshots_Empty(t *testing.T) {
	cfg, _ := setupSnapshotVault(t)

	snaps, err := vault.ListSnapshots(cfg)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}
}

func TestSaveSnapshot_NotInitialised(t *testing.T) {
	tmp := t.TempDir()
	cfg := vault.DefaultConfig(tmp)

	err := vault.SaveSnapshot(cfg, "test")
	if err == nil {
		t.Fatal("expected error for uninitialised vault")
	}
}

func TestSaveSnapshot_MultipleSnapshots(t *testing.T) {
	cfg, _ := setupSnapshotVault(t)

	for _, label := range []string{"snap-1", "snap-2", "snap-3"} {
		if err := vault.SaveSnapshot(cfg, label); err != nil {
			t.Fatalf("SaveSnapshot(%s): %v", label, err)
		}
	}

	snaps, err := vault.ListSnapshots(cfg)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(snaps))
	}
}

func TestSnapshotDiff(t *testing.T) {
	a := vault.Snapshot{
		Keys: map[string]string{"A": "hash1", "B": "hash2"},
	}
	b := vault.Snapshot{
		Keys: map[string]string{"A": "hash1", "B": "hashX", "C": "hash3"},
	}

	lines := vault.SnapshotDiff(a, b)
	found := map[string]bool{}
	for _, l := range lines {
		found[l] = true
	}
	if !found["~ B (changed)"] {
		t.Error("expected B changed")
	}
	if !found["+ C (added)"] {
		t.Error("expected C added")
	}
	_ = os.Getenv("CI") // suppress unused import lint
}
