package vault

// RemoveRecipientAndRekey removes a recipient by their public key and
// immediately rekeys the encrypted file so the removed party can no
// longer decrypt future versions of the vault.
//
// It is a convenience wrapper around RemoveRecipient + Rekey that
// ensures both operations succeed atomically from the caller's
// perspective: if either step fails the function returns an error and
// the vault is left in whatever state the failing step produced.
func RemoveRecipientAndRekey(cfg Config, pubkey string) error {
	if err := RemoveRecipient(cfg, pubkey); err != nil {
		return err
	}
	return Rekey(cfg)
}
