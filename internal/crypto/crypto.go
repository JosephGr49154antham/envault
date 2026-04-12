package crypto

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

// Encrypt encrypts plaintext using the provided age recipients.
func Encrypt(plaintext []byte, recipients []age.Recipient) ([]byte, error) {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipients...)
	if err != nil {
		return nil, fmt.Errorf("initializing age encryption: %w", err)
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("writing plaintext: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("finalizing encryption: %w", err)
	}
	return buf.Bytes(), nil
}

// Decrypt decrypts ciphertext using the provided age identities.
func Decrypt(ciphertext []byte, identities []age.Identity) ([]byte, error) {
	r, err := age.Decrypt(bytes.NewReader(ciphertext), identities...)
	if err != nil {
		return nil, fmt.Errorf("initializing age decryption: %w", err)
	}
	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading decrypted data: %w", err)
	}
	return plaintext, nil
}

// EncryptFile encrypts the contents of src and writes ciphertext to dst.
func EncryptFile(src io.Reader, dst io.Writer, recipients []age.Recipient) error {
	w, err := age.Encrypt(dst, recipients...)
	if err != nil {
		return fmt.Errorf("initializing age encryption: %w", err)
	}
	if _, err := io.Copy(w, src); err != nil {
		return fmt.Errorf("encrypting data: %w", err)
	}
	return w.Close()
}

// DecryptFile decrypts ciphertext from src and writes plaintext to dst.
func DecryptFile(src io.Reader, dst io.Writer, identities []age.Identity) error {
	r, err := age.Decrypt(src, identities...)
	if err != nil {
		return fmt.Errorf("initializing age decryption: %w", err)
	}
	_, err = io.Copy(dst, r)
	return err
}
