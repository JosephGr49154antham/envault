package vault

// RecipientInfo holds display information about a single recipient.
type RecipientInfo struct {
	// Label is the short human-readable name derived from the key comment or
	// the last path segment of the public key string.
	Label string
	// PublicKey is the full age public-key string (age1…).
	PublicKey string
}

// ListRecipientsInfo returns structured RecipientInfo entries for every
// public key stored in the vault's recipients file.  It is similar to
// ListRecipients but returns richer data suitable for tabular output.
func ListRecipientsInfo(cfg Config) ([]RecipientInfo, error) {
	if !IsInitialised(cfg) {
		return nil, ErrNotInitialised
	}

	pubs, err := LoadRecipientKeys(cfg)
	if err != nil {
		return nil, err
	}

	infos := make([]RecipientInfo, 0, len(pubs))
	for _, pub := range pubs {
		infos = append(infos, RecipientInfo{
			Label:     lastSegmentOf(pub),
			PublicKey: pub,
		})
	}
	return infos, nil
}

// LoadRecipientKeys reads the recipients file and returns only the raw
// public-key strings, skipping blank lines and comments.
func LoadRecipientKeys(cfg Config) ([]string, error) {
	recips, err := loadRecipientsFromFile(cfg.RecipientsFile)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(recips))
	for _, r := range recips {
		keys = append(keys, r.String())
	}
	return keys, nil
}
