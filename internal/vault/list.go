package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EnvFileInfo holds metadata about a tracked env file in the vault.
type EnvFileInfo struct {
	Name      string
	PlainPath string
	EncPath   string
	HasPlain  bool
	HasEnc    bool
	EncMtime  time.Time
}

// List returns metadata for all env files tracked in the vault.
// It scans the encrypted directory for .env.age files and resolves
// their corresponding plaintext counterparts.
func List(cfg Config) ([]EnvFileInfo, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	entries, err := os.ReadDir(cfg.EncryptedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []EnvFileInfo{}, nil
		}
		return nil, fmt.Errorf("reading encrypted dir: %w", err)
	}

	var infos []EnvFileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".age") {
			continue
		}

		baseName := strings.TrimSuffix(name, ".age")
		encPath := filepath.Join(cfg.EncryptedDir, name)
		plainPath := baseName

		fi, statErr := e.Info()
		var encMtime time.Time
		if statErr == nil {
			encMtime = fi.ModTime()
		}

		_, plainErr := os.Stat(plainPath)
		_, encErr := os.Stat(encPath)

		infos = append(infos, EnvFileInfo{
			Name:      baseName,
			PlainPath: plainPath,
			EncPath:   encPath,
			HasPlain:  plainErr == nil,
			HasEnc:    encErr == nil,
			EncMtime:  encMtime,
		})
	}
	return infos, nil
}

// Find returns the EnvFileInfo for the named env file, or an error if it is
// not tracked in the vault. name should match the base filename (e.g. ".env").
func Find(cfg Config, name string) (EnvFileInfo, error) {
	infos, err := List(cfg)
	if err != nil {
		return EnvFileInfo{}, err
	}
	for _, info := range infos {
		if info.Name == name {
			return info, nil
		}
	}
	return EnvFileInfo{}, fmt.Errorf("no vault entry found for %q", name)
}
