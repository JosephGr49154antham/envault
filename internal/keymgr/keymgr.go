// Package keymgr handles generation and persistence of age key pairs
// used for encrypting and decrypting .env files.
package keymgr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"
)

const (
	DefaultKeyDir  = ".envault"
	DefaultKeyFile = "identity.txt"
)

// KeyPair holds an age X25519 identity (private key) and its recipient (public key).
type KeyPair struct {
	Identity  *age.X25519Identity
	Recipient *age.X25519Recipient
}

// GenerateKeyPair creates a new age X25519 key pair.
func GenerateKeyPair() (*KeyPair, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("generating key pair: %w", err)
	}
	return &KeyPair{
		Identity:  identity,
		Recipient: identity.Recipient(),
	}, nil
}

// SaveIdentity writes the private key to disk at the given path.
// The file is created with 0600 permissions.
func SaveIdentity(identity *age.X25519Identity, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating key directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening key file: %w", err)
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "# envault age identity\n%s\n", identity.String())
	return err
}

// LoadIdentity reads an age X25519 private key from the given path.
func LoadIdentity(path string) (*age.X25519Identity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("identity file not found at %s; run 'envault init' first", path)
		}
		return nil, fmt.Errorf("reading identity file: %w", err)
	}
	identities, err := age.ParseIdentities(bytesReader(data))
	if err != nil {
		return nil, fmt.Errorf("parsing identity: %w", err)
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("no identities found in %s", path)
	}
	id, ok := identities[0].(*age.X25519Identity)
	if !ok {
		return nil, fmt.Errorf("unexpected identity type")
	}
	return id, nil
}

// DefaultIdentityPath returns the default path for the identity file
// relative to the user's home directory.
func DefaultIdentityPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, DefaultKeyDir, DefaultKeyFile), nil
}

// IdentityExists reports whether an identity file exists at the given path.
func IdentityExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("checking identity file: %w", err)
}
