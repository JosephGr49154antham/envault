package recipients_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/recipients"
)

func TestAddAndLoadRecipients(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recipients")

	// Generate two key pairs and add their public keys.
	for i := 0; i < 2; i++ {
		id, err := keymgr.GenerateKeyPair()
		if err != nil {
			t.Fatalf("GenerateKeyPair: %v", err)
		}
		pubkey := id.Recipient().String()
		if err := recipients.AddRecipient(path, pubkey); err != nil {
			t.Fatalf("AddRecipient: %v", err)
		}
	}

	loaded, err := recipients.LoadRecipients(path)
	if err != nil {
		t.Fatalf("LoadRecipients: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(loaded))
	}
}

func TestLoadRecipients_NotFound(t *testing.T) {
	_, err := recipients.LoadRecipients("/nonexistent/path/recipients")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadRecipients_SkipsCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recipients")

	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	content := "# This is a comment\n\n" + id.Recipient().String() + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	loaded, err := recipients.LoadRecipients(path)
	if err != nil {
		t.Fatalf("LoadRecipients: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("expected 1 recipient, got %d", len(loaded))
	}
}

func TestAddRecipient_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recipients")

	err := recipients.AddRecipient(path, "not-a-valid-age-key")
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
}
