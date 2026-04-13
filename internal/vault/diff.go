package vault

import (
	"fmt"
	"os"
	"strings"
)

// DiffResult holds the comparison between a plain and encrypted env file.
type DiffResult struct {
	PlainPath     string
	EncryptedPath string
	OnlyInPlain   []string
	OnlyInEnc     []string
	Changed       []string
	Unchanged     []string
}

// HasDiff returns true if there are any differences between the two files.
func (d DiffResult) HasDiff() bool {
	return len(d.OnlyInPlain) > 0 || len(d.OnlyInEnc) > 0 || len(d.Changed) > 0
}

// Diff compares a plain .env file with its decrypted counterpart and returns
// a DiffResult describing which keys were added, removed, or changed.
func Diff(cfg Config, plainPath string) (DiffResult, error) {
	if !IsInitialised(cfg) {
		return DiffResult{}, fmt.Errorf("vault not initialised; run 'envault init' first")
	}

	encPath := cfg.EncryptedFile
	result := DiffResult{
		PlainPath:     plainPath,
		EncryptedPath: encPath,
	}

	plainVars, err := parseEnvFile(plainPath)
	if err != nil {
		return DiffResult{}, fmt.Errorf("reading plain file: %w", err)
	}

	if _, err := os.Stat(encPath); os.IsNotExist(err) {
		// Everything in plain is "only in plain"
		for k := range plainVars {
			result.OnlyInPlain = append(result.OnlyInPlain, k)
		}
		return result, nil
	}

	identity, err := loadVaultIdentity(cfg)
	if err != nil {
		return DiffResult{}, err
	}

	decrypted, err := decryptToMap(encPath, identity)
	if err != nil {
		return DiffResult{}, fmt.Errorf("decrypting vault file: %w", err)
	}

	for k, v := range plainVars {
		if ev, ok := decrypted[k]; !ok {
			result.OnlyInPlain = append(result.OnlyInPlain, k)
		} else if ev != v {
			result.Changed = append(result.Changed, k)
		} else {
			result.Unchanged = append(result.Unchanged, k)
		}
	}

	for k := range decrypted {
		if _, ok := plainVars[k]; !ok {
			result.OnlyInEnc = append(result.OnlyInEnc, k)
		}
	}

	return result, nil
}

// parseEnvFile reads a KEY=VALUE env file and returns a map of entries,
// skipping blank lines and comments.
func parseEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result, nil
}
