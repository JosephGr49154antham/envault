package vault

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/recipients"
)

// Import encrypts a plain .env file from an arbitrary source path and writes
// the encrypted output into the vault, optionally overwriting an existing
// encrypted file.
func Import(cfg Config, src string, overwrite bool) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", src)
	}

	recips, err := recipients.LoadRecipients(cfg.RecipientsFile)
	if err != nil {
		return fmt.Errorf("load recipients: %w", err)
	}
	if len(recips) == 0 {
		return fmt.Errorf("no recipients configured; add at least one with 'envault add-recipient'")
	}

	dst := cfg.EncryptedFile
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("encrypted file already exists at %s; use --overwrite to replace it", dst)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}

	if err := crypto.EncryptFile(src, dst, recips); err != nil {
		return fmt.Errorf("encrypt file: %w", err)
	}

	return nil
}
