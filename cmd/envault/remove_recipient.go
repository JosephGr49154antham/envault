package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/envault/envault/internal/vault"
)

func runRemoveRecipient(args []string) error {
	fs := flag.NewFlagSet("remove-recipient", flag.ContinueOnError)
	rekey := fs.Bool("rekey", true, "rekey the vault after removing the recipient (recommended)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: envault remove-recipient [--rekey=false] <public-key>")
		return fmt.Errorf("public key argument required")
	}

	pubkey := fs.Arg(0)
	cfg := vault.DefaultConfig()

	if *rekey {
		if err := vault.RemoveRecipientAndRekey(cfg, pubkey); err != nil {
			return fmt.Errorf("remove-recipient: %w", err)
		}
		fmt.Printf("Recipient removed and vault rekeyed.\n")
		return nil
	}

	if err := vault.RemoveRecipient(cfg, pubkey); err != nil {
		return fmt.Errorf("remove-recipient: %w", err)
	}
	fmt.Printf("Recipient removed (vault not rekeyed — run 'envault rekey' to rotate encryption).\n")
	return nil
}
