package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const lockFileName = ".envault.lock"

// LockInfo holds metadata written into the lock file.
type LockInfo struct {
	User      string    `json:"user"`
	LockedAt  time.Time `json:"locked_at"`
	MachineName string  `json:"machine"`
}

// lockFilePath returns the path to the lock file inside the vault dir.
func lockFilePath(cfg Config) string {
	return filepath.Join(cfg.VaultDir, lockFileName)
}

// Lock creates a lock file in the vault directory, preventing concurrent
// push/pull operations. Returns an error if the vault is already locked.
func Lock(cfg Config) error {
	if !IsInitialised(cfg) {
		return errors.New("vault is not initialised; run 'envault init' first")
	}

	path := lockFilePath(cfg)
	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		return fmt.Errorf("vault is already locked: %s", string(data))
	}

	user, _ := CurrentGitUser()
	hostname, _ := os.Hostname()

	contents := fmt.Sprintf("user=%s machine=%s locked_at=%s\n", user, hostname, time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	return nil
}

// Unlock removes the lock file from the vault directory.
func Unlock(cfg Config) error {
	if !IsInitialised(cfg) {
		return errors.New("vault is not initialised; run 'envault init' first")
	}

	path := lockFilePath(cfg)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("vault is not locked")
		}
		return fmt.Errorf("failed to remove lock file: %w", err)
	}
	return nil
}

// IsLocked reports whether the vault currently has an active lock file.
func IsLocked(cfg Config) bool {
	_, err := os.Stat(lockFilePath(cfg))
	return err == nil
}

// LockStatus returns a human-readable lock status string.
func LockStatus(cfg Config) string {
	path := lockFilePath(cfg)
	data, err := os.ReadFile(path)
	if err != nil {
		return "unlocked"
	}
	return "locked: " + string(data)
}
