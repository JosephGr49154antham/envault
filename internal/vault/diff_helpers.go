package vault

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nicholasgasior/envault/internal/crypto"
	"github.com/nicholasgasior/envault/internal/keymgr"
	age "filippo.io/age"
)

// loadVaultIdentity loads the age identity from the default path.
func loadVaultIdentity(cfg Config) (*age.X25519Identity, error) {
	identityPath := keymgr.DefaultIdentityPath()
	identity, err := keymgr.LoadIdentity(identityPath)
	if err != nil {
		return nil, fmt.Errorf("loading identity from %s: %w", identityPath, err)
	}
	return identity, nil
}

// decryptToMap decrypts an age-encrypted env file and parses it into a map.
func decryptToMap(encPath string, identity *age.X25519Identity) (map[string]string, error) {
	var buf bytes.Buffer
	if err := crypto.DecryptFile(encPath, &buf, identity); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, line := range strings.Split(buf.String(), "\n") {
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
