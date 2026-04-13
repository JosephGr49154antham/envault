package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

// Clone copies a decrypted .env file from one named profile to another within
// the vault directory. The source file must exist as a plaintext .env file.
// If overwrite is false and the destination already exists, an error is returned.
func Clone(cfg Config, srcName, dstName string, overwrite bool) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised: run 'envault init' first")
	}

	srcPath := filepath.Join(cfg.VaultDir, srcName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("source env file not found: %s", srcPath)
	}

	dstPath := filepath.Join(cfg.VaultDir, dstName)
	if !overwrite {
		if _, err := os.Stat(dstPath); err == nil {
			return fmt.Errorf("destination already exists: %s (use --overwrite to replace)", dstPath)
		}
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	if err := os.WriteFile(dstPath, data, 0600); err != nil {
		return fmt.Errorf("writing destination file: %w", err)
	}

	return nil
}
