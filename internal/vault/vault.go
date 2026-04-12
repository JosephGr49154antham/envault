package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultVaultDir  = ".envault"
	RecipientsFile   = "recipients.txt"
	EncryptedEnvFile = ".env.age"
)

// Config holds the vault configuration for a project.
type Config struct {
	// VaultDir is the directory where vault metadata is stored.
	VaultDir string
	// RecipientsPath is the path to the recipients file.
	RecipientsPath string
	// EncryptedPath is the path to the encrypted env file.
	EncryptedPath string
	// PlaintextPath is the path to the plaintext .env file.
	PlaintextPath string
}

// DefaultConfig returns a Config rooted at the given project directory.
func DefaultConfig(projectDir string) *Config {
	vaultDir := filepath.Join(projectDir, DefaultVaultDir)
	return &Config{
		VaultDir:       vaultDir,
		RecipientsPath: filepath.Join(vaultDir, RecipientsFile),
		EncryptedPath:  filepath.Join(vaultDir, EncryptedEnvFile),
		PlaintextPath:  filepath.Join(projectDir, ".env"),
	}
}

// Init initialises the vault directory structure for a project.
// It creates the vault directory if it does not already exist.
func Init(projectDir string) (*Config, error) {
	cfg := DefaultConfig(projectDir)

	if err := os.MkdirAll(cfg.VaultDir, 0700); err != nil {
		return nil, fmt.Errorf("vault: create vault dir: %w", err)
	}

	// Create an empty recipients file if one does not exist.
	if _, err := os.Stat(cfg.RecipientsPath); errors.Is(err, os.ErrNotExist) {
		f, err := os.OpenFile(cfg.RecipientsPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("vault: create recipients file: %w", err)
		}
		f.Close()
	}

	return cfg, nil
}

// IsInitialised reports whether the vault directory exists at projectDir.
func IsInitialised(projectDir string) bool {
	cfg := DefaultConfig(projectDir)
	_, err := os.Stat(cfg.VaultDir)
	return err == nil
}
