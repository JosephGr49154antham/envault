package vault

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// RemoveRecipient removes a recipient public key from the vault's recipients file
// and re-encrypts the vault secrets so the removed recipient can no longer decrypt them.
// It returns an error if the vault is not initialised, the key is not found, or re-encryption fails.
func RemoveRecipient(cfg Config, pubKey string) error {
	if !IsInitialised(cfg) {
		return errors.New("vault is not initialised; run 'envault init' first")
	}

	pubKey = strings.TrimSpace(pubKey)
	if pubKey == "" {
		return errors.New("public key must not be empty")
	}

	data, err := os.ReadFile(cfg.RecipientsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("recipients file not found: %s", cfg.RecipientsFile)
		}
		return fmt.Errorf("reading recipients file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	updated := make([]string, 0, len(lines))
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == pubKey {
			found = true
			continue // skip this recipient
		}
		updated = append(updated, line)
	}

	if !found {
		return fmt.Errorf("recipient not found in %s", cfg.RecipientsFile)
	}

	// Trim trailing blank lines but preserve a final newline
	result := strings.TrimRight(strings.Join(updated, "\n"), "\n")
	if result != "" {
		result += "\n"
	}

	if err := os.WriteFile(cfg.RecipientsFile, []byte(result), 0o644); err != nil {
		return fmt.Errorf("writing recipients file: %w", err)
	}

	// Re-key so removed recipient can no longer decrypt vault secrets.
	if err := Rekey(cfg); err != nil {
		return fmt.Errorf("re-keying after recipient removal: %w", err)
	}

	return nil
}
