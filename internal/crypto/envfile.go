package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"
)

const encryptedExtension = ".age"

// EncryptEnvFile reads a .env file at src, encrypts it, and writes the
// result to dst (e.g. ".env.age"). If dst is empty, it defaults to src+".age".
func EncryptEnvFile(src string, dst string, recipients []age.Recipient) error {
	if dst == "" {
		dst = src + encryptedExtension
	}

	plaintext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading env file %q: %w", src, err)
	}

	ciphertext, err := Encrypt(plaintext, recipients)
	if err != nil {
		return fmt.Errorf("encrypting env file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(dst, ciphertext, 0o600); err != nil {
		return fmt.Errorf("writing encrypted file %q: %w", dst, err)
	}
	return nil
}

// DecryptEnvFile reads an encrypted .env file at src, decrypts it, and writes
// the plaintext to dst. If dst is empty, it strips the ".age" extension from src.
func DecryptEnvFile(src string, dst string, identities []age.Identity) error {
	if dst == "" {
		if filepath.Ext(src) == encryptedExtension {
			dst = src[:len(src)-len(encryptedExtension)]
		} else {
			dst = src + ".decrypted"
		}
	}

	ciphertext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading encrypted file %q: %w", src, err)
	}

	plaintext, err := Decrypt(ciphertext, identities)
	if err != nil {
		return fmt.Errorf("decrypting env file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(dst, plaintext, 0o600); err != nil {
		return fmt.Errorf("writing decrypted file %q: %w", dst, err)
	}
	return nil
}
