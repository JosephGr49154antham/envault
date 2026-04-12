package keymgr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envault/envault/internal/keymgr"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}
	if kp.Identity == nil {
		t.Fatal("expected non-nil Identity")
	}
	if kp.Recipient == nil {
		t.Fatal("expected non-nil Recipient")
	}
	if kp.Identity.Recipient().String() != kp.Recipient.String() {
		t.Error("recipient derived from identity does not match stored recipient")
	}
}

func TestSaveAndLoadIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.txt")

	kp, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	if err := keymgr.SaveIdentity(kp.Identity, path); err != nil {
		t.Fatalf("SaveIdentity() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat identity file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}

	loaded, err := keymgr.LoadIdentity(path)
	if err != nil {
		t.Fatalf("LoadIdentity() error = %v", err)
	}
	if loaded.String() != kp.Identity.String() {
		t.Errorf("loaded identity %q != original %q", loaded.String(), kp.Identity.String())
	}
}

func TestLoadIdentity_NotFound(t *testing.T) {
	_, err := keymgr.LoadIdentity("/nonexistent/path/identity.txt")
	if err == nil {
		t.Fatal("expected error for missing identity file, got nil")
	}
}

func TestDefaultIdentityPath(t *testing.T) {
	path, err := keymgr.DefaultIdentityPath()
	if err != nil {
		t.Fatalf("DefaultIdentityPath() error = %v", err)
	}
	if path == "" {
		t.Error("expected non-empty default identity path")
	}
}
