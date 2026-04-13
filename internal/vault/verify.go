package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// VerifyResult holds the outcome of verifying a single encrypted file.
type VerifyResult struct {
	File    string
	Valid   bool
	Checksum string
	Error   error
}

// Verify checks the integrity of all encrypted .env files in the vault by
// attempting to decrypt them and computing their SHA-256 checksums.
// It returns one VerifyResult per encrypted file found.
func Verify(cfg Config) ([]VerifyResult, error) {
	if !IsInitialised(cfg) {
		return nil, errors.New("vault is not initialised; run 'envault init' first")
	}

	identity, err := loadVaultIdentity(cfg)
	if err != nil {
		return nil, fmt.Errorf("load identity: %w", err)
	}

	entries, err := os.ReadDir(cfg.VaultDir)
	if err != nil {
		return nil, fmt.Errorf("read vault dir: %w", err)
	}

	var results []VerifyResult
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".age" {
			continue
		}

		path := filepath.Join(cfg.VaultDir, e.Name())
		result := VerifyResult{File: e.Name()}

		plaintext, err := decryptToMap(path, identity)
		if err != nil {
			result.Valid = false
			result.Error = err
			results = append(results, result)
			continue
		}

		checksum, err := checksumFile(path)
		if err != nil {
			result.Valid = false
			result.Error = fmt.Errorf("checksum: %w", err)
			results = append(results, result)
			continue
		}

		_ = plaintext // decryption succeeded; content is valid
		result.Valid = true
		result.Checksum = checksum
		results = append(results, result)
	}

	return results, nil
}

// checksumFile returns the hex-encoded SHA-256 digest of the file at path.
func checksumFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
