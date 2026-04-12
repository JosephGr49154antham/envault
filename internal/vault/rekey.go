package vault

import (
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/recipients"
)

// Rekey re-encrypts the vault's encrypted .env file for the current set of
// recipients. This should be called after adding or removing a recipient so
// that the ciphertext reflects the latest recipient list.
//
// The caller must supply an identity that is already authorised (i.e. present
// in the recipients file) so that the existing ciphertext can be decrypted
// before it is re-encrypted for the new recipient set.
func Rekey(cfg Config, identityPath string) error {
	// Ensure the vault has been initialised.
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	// Load the caller's identity so we can decrypt the existing ciphertext.
	id, err := keymgr.LoadIdentity(identityPath)
	if err != nil {
		return fmt.Errorf("load identity: %w", err)
	}

	// Check that the encrypted file actually exists before we attempt anything.
	if _, err := os.Stat(cfg.EncryptedFile); os.IsNotExist(err) {
		return fmt.Errorf("encrypted file not found at %s; run 'envault push' first", cfg.EncryptedFile)
	}

	// Read and decrypt the existing ciphertext.
	ciphertext, err := os.ReadFile(cfg.EncryptedFile)
	if err != nil {
		return fmt.Errorf("read encrypted file: %w", err)
	}

	plaintext, err := crypto.Decrypt(ciphertext, id)
	if err != nil {
		return fmt.Errorf("decrypt existing vault (wrong identity?): %w", err)
	}

	// Load the current recipient list.
	recips, err := recipients.LoadRecipients(cfg.RecipientsFile)
	if err != nil {
		return fmt.Errorf("load recipients: %w", err)
	}
	if len(recips) == 0 {
		return fmt.Errorf("no recipients found in %s; add at least one recipient before rekeying", cfg.RecipientsFile)
	}

	// Re-encrypt the plaintext for the updated recipient set.
	newCiphertext, err := crypto.Encrypt(plaintext, recips)
	if err != nil {
		return fmt.Errorf("re-encrypt vault: %w", err)
	}

	// Atomically overwrite the encrypted file.
	if err := os.WriteFile(cfg.EncryptedFile, newCiphertext, 0o644); err != nil {
		return fmt.Errorf("write rekeyed vault: %w", err)
	}

	return nil
}
