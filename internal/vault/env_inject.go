package vault

import (
	"fmt"
	"os"
	"strings"

	"github.com/nicholasgasior/envault/internal/crypto"
)

// InjectOptions controls the behaviour of Inject.
type InjectOptions struct {
	// Src is the encrypted .env.age file to decrypt and inject.
	Src string
	// Overwrite, if true, allows existing env vars in the process environment
	// to be overwritten by values from the vault.
	Overwrite bool
	// DryRun prints the variables that would be set without modifying the
	// environment.
	DryRun bool
}

// InjectResult summarises what Inject did.
type InjectResult struct {
	Set     []string
	Skipped []string
}

// Inject decrypts the given encrypted env file and injects the key=value
// pairs into the current process environment. It is primarily useful for
// sub-command execution scenarios (e.g. envault inject -- myserver).
func Inject(cfg Config, opts InjectOptions) (*InjectResult, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	id, err := loadVaultIdentity(cfg)
	if err != nil {
		return nil, fmt.Errorf("load identity: %w", err)
	}

	src := opts.Src
	if src == "" {
		src = cfg.EncryptedFile
	}

	plaintext, err := crypto.DecryptFile(src, id)
	if err != nil {
		return nil, fmt.Errorf("decrypt %s: %w", src, err)
	}

	result := &InjectResult{}
	for _, line := range strings.Split(string(plaintext), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		_, exists := os.LookupEnv(key)
		if exists && !opts.Overwrite {
			result.Skipped = append(result.Skipped, key)
			continue
		}
		if !opts.DryRun {
			if err := os.Setenv(key, val); err != nil {
				return nil, fmt.Errorf("setenv %s: %w", key, err)
			}
		}
		result.Set = append(result.Set, key)
	}
	return result, nil
}
