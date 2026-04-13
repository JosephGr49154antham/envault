package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cqroot/envault/internal/vault"
)

func runVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault verify")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Verify the integrity of all encrypted .env files in the vault.")
		fmt.Fprintln(os.Stderr, "Each file is decrypted using your identity and its SHA-256 checksum is printed.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	results, err := vault.Verify(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
	("No encrypted files found in vaultValid := true
	for _, := range results {
	t	fmt.Printf("  ✓  %-30s  sha256:%s\n", r.File, r.Checksum)
		} else {
			fmt.Printf("  ✗  %-30s  ERROR: %v\n", r.File, r.Error)
			allValid = false
		}
	}

	if !allValid {
		fmt.Fprintln(os.Stderr, "\nOne or more files failed verification.")
		os.Exit(1)
	}

	fmt.Println("\nAll files verified successfully.")
}
