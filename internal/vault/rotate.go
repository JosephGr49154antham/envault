package vault

import (
	"fmt"
	"os"
	"time"
)

// RotateOptions configures the key rotation behaviour.
type RotateOptions struct {
	// BackupDir is where the old encrypted file is backed up before rotation.
	// Defaults to <VaultDir>/backups if empty.
	BackupDir string
}

// Rotate re-encrypts the vault env file after a recipient list change.
// It first backs up the current encrypted file, then calls Rekey so that
// only the current recipients can decrypt the new ciphertext.
//
// The backup is written to <BackupDir>/<timestamp>_<encryptedFilename>.
func Rotate(cfg Config, opts RotateOptions) error {
	if !IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	backupDir := opts.BackupDir
	if backupDir == "" {
		backupDir = cfg.VaultDir + "/backups"
	}

	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	// Only back up if an encrypted file already exists.
	if _, err := os.Stat(cfg.EncryptedFile); err == nil {
		timestamp := time.Now().UTC().Format("20060102T150405Z")
		backupName := fmt.Sprintf("%s_%s", timestamp, lastSegment(cfg.EncryptedFile))
		backupPath := backupDir + "/" + backupName

		data, readErr := os.ReadFile(cfg.EncryptedFile)
		if readErr != nil {
			return fmt.Errorf("read encrypted file for backup: %w", readErr)
		}
		if writeErr := os.WriteFile(backupPath, data, 0o600); writeErr != nil {
			return fmt.Errorf("write backup: %w", writeErr)
		}
	}

	return Rekey(cfg)
}

// lastSegment returns the last path segment of p (reuses recipients helper).
func lastSegment(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}
