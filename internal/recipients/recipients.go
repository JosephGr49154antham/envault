package recipients

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"filippo.io/age"
)

const DefaultRecipientsFile = ".envault/recipients"

// LoadRecipients reads public keys (age recipients) from a file,
// one per line, ignoring blank lines and comments starting with '#'.
func LoadRecipients(path string) ([]age.Recipient, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("recipients file not found: %s", path)
		}
		return nil, fmt.Errorf("open recipients file: %w", err)
	}
	defer f.Close()

	var recipients []age.Recipient
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		r, err := age.ParseX25519Recipient(line)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient on line %d: %w", lineNum, err)
		}
		recipients = append(recipients, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading recipients file: %w", err)
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("no valid recipients found in %s", path)
	}
	return recipients, nil
}

// AddRecipient appends a public key to the recipients file, creating it if needed.
func AddRecipient(path, pubkey string) error {
	// Validate the public key before writing.
	if _, err := age.ParseX25519Recipient(pubkey); err != nil {
		return fmt.Errorf("invalid age public key: %w", err)
	}

	if err := os.MkdirAll(strings.TrimSuffix(path, "/"+lastSegment(path)), 0700); err != nil {
		return fmt.Errorf("create recipients dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open recipients file for writing: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, pubkey)
	return err
}

func lastSegment(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
