// Package vault provides core vault operations for envault.
package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/envault/envault/internal/crypto"
)

// ExportOptions configures an Export operation.
type ExportOptions struct {
	// OutputPath is the destination file. If empty, defaults to
	// <plaintext-basename>-export-<timestamp>.env next to the vault dir.
	OutputPath string
	// Overwrite allows clobbering an existing file.
	Overwrite bool
}

// Export decrypts the vault's encrypted env file and writes the plaintext to
// the path described by opts. It returns the resolved output path.
func Export(cfg Config, opts ExportOptions) (string, error) {
	if !IsInitialised(cfg) {
		return "", fmt.Errorf("vault is not initialised; run `envault init` first")
	}

	if _, err := os.Stat(cfg.EncryptedFile); os.IsNotExist(err) {
		return "", fmt.Errorf("encrypted file not found: %s", cfg.EncryptedFile)
	}

	identity, err := loadVaultIdentity(cfg)
	if err != nil {
		return "", fmt.Errorf("load identity: %w", err)
	}

	out := opts.OutputPath
	if out == "" {
		base := filepath.Base(cfg.PlainFile)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		timestamp := time.Now().Format("20060102-150405")
		out = filepath.Join(filepath.Dir(cfg.VaultDir), fmt.Sprintf("%s-export-%s%s", name, timestamp, ext))
	}

	if !opts.Overwrite {
		if _, err := os.Stat(out); err == nil {
			return "", fmt.Errorf("output file already exists: %s (use --overwrite to replace)", out)
		}
	}

	if err := crypto.DecryptFile(cfg.EncryptedFile, out, identity); err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return out, nil
}
