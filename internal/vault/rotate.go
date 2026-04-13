package vault

import (
	"fmt"
	"os"
	"path/filepath"
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
		if err := backupEncryptedFile(cfg.EncryptedFile, backupDir); err != nil {
			return err
		}
	}

	return Rekey(cfg)
}

// backupEncryptedFile copies src into backupDir with a timestamp prefix so
// that multiple rotations do not overwrite each other.
func backupEncryptedFile(src, backupDir string) error {
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	backupName := fmt.Sprintf("%s_%s", timestamp, filepath.Base(src))
	backupPath := filepath.Join(backupDir, backupName)

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read encrypted file for backup: %w", err)
	}
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}
	return nil
}
