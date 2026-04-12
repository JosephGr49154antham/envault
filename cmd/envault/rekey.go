package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/envault/envault/internal/keymgr"
	"github.com/envault/envault/internal/vault"
)

func runRekey(args []string) error {
	fs := flag.NewFlagSet("rekey", flag.ContinueOnError)
	root := fs.String("root", ".", "vault root directory")
	keyPath := fs.String("identity", keymgr.DefaultIdentityPath(), "path to age identity file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig(*root)

	if !vault.IsInitialised(cfg) {
		return fmt.Errorf("vault is not initialised in %s; run `envault init` first", *root)
	}

	id, err := keymgr.LoadIdentity(*keyPath)
	if err != nil {
		return fmt.Errorf("load identity from %s: %w", *keyPath, err)
	}

	if err := vault.Rekey(cfg, id); err != nil {
		return fmt.Errorf("rekey: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Vault rekeyed successfully using recipients in %s\n", cfg.RecipientsFile)
	return nil
}
