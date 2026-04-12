package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileStatus represents the sync status of a single env file.
type FileStatus struct {
	PlainFile     string
	EncryptedFile string
	PlainExists   bool
	EncExists     bool
	PlainModTime  time.Time
	EncModTime    time.Time
	InSync        bool
	Stale         bool // plain is newer than encrypted
}

// Status returns the sync status of all tracked env files in the vault.
func Status(cfg Config) ([]FileStatus, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	entries, err := os.ReadDir(cfg.VaultDir)
	if err != nil {
		return nil, fmt.Errorf("reading vault dir: %w", err)
	}

	var statuses []FileStatus

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".age" {
			continue
		}

		base := e.Name()[:len(e.Name())-len(".age")]
		encPath := filepath.Join(cfg.VaultDir, e.Name())
		plainPath := base // relative to working dir

		fs := FileStatus{
			PlainFile:     plainPath,
			EncryptedFile: encPath,
			EncExists:     true,
		}

		encInfo, err := os.Stat(encPath)
		if err == nil {
			fs.EncModTime = encInfo.ModTime()
		}

		plainInfo, err := os.Stat(plainPath)
		if err == nil {
			fs.PlainExists = true
			fs.PlainModTime = plainInfo.ModTime()
		}

		if fs.PlainExists && fs.EncExists {
			fs.Stale = fs.PlainModTime.After(fs.EncModTime)
			fs.InSync = !fs.Stale
		}

		statuses = append(statuses, fs)
	}

	return statuses, nil
}
