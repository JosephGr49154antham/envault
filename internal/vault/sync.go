package vault

import (
	"fmt"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/recipients"
	"filippo.io/age"
)

// Push encrypts the plaintext .env file and writes the result to the vault.
// It reads the current recipients list from cfg.RecipientsPath.
func Push(cfg *Config, identity age.Identity) error {
	recips, err := recipients.LoadRecipients(cfg.RecipientsPath)
	if err != nil {
		return fmt.Errorf("vault push: load recipients: %w", err)
	}
	if len(recips) == 0 {
		return fmt.Errorf("vault push: no recipients configured — add at least one recipient first")
	}

	if err := crypto.EncryptEnvFile(cfg.PlaintextPath, cfg.EncryptedPath, recips); err != nil {
		return fmt.Errorf("vault push: encrypt: %w", err)
	}
	return nil
}

// Pull decrypts the encrypted vault file and writes the plaintext .env file.
func Pull(cfg *Config, identity age.Identity) error {
	if err := crypto.DecryptEnvFile(cfg.EncryptedPath, cfg.PlaintextPath, identity); err != nil {
		return fmt.Errorf("vault pull: decrypt: %w", err)
	}
	return nil
}
