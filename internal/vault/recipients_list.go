package vault

import (
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/recipients"
)

// RecipientEntry holds display information about a single recipient.
type RecipientEntry struct {
	// Label is the short identifier derived from the key (last segment).
	Label string
	// PublicKey is the full age public key string.
	PublicKey string
}

// ListRecipients returns all recipients registered in the vault.
// It returns an error if the vault is not initialised or the recipients
// file cannot be read.
func ListRecipients(cfg Config) ([]RecipientEntry, error) {
	if !IsInitialised(cfg) {
		return nil, fmt.Errorf("vault is not initialised; run 'envault init' first")
	}

	pubs, err := recipients.LoadRecipients(cfg.RecipientsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []RecipientEntry{}, nil
		}
		return nil, fmt.Errorf("load recipients: %w", err)
	}

	entries := make([]RecipientEntry, 0, len(pubs))
	for _, r := range pubs {
		entries = append(entries, RecipientEntry{
			Label:     lastSegmentOf(r.String()),
			PublicKey: r.String(),
		})
	}
	return entries, nil
}

// lastSegmentOf returns the final colon-separated segment of an age
// public-key string, used as a short human-readable label.
func lastSegmentOf(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return key[i+1:]
		}
	}
	return key
}
