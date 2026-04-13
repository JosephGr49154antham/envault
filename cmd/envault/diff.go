package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	plainFile := fs.String("f", ".env", "path to the plain .env file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig()
	result, err := vault.Diff(cfg, *plainFile)
	if err != nil {
		return err
	}

	if !result.HasDiff() {
		fmt.Println("✓ env file is in sync with vault")
		return nil
	}

	fmt.Printf("Diff between %s and %s:\n\n", result.PlainPath, result.EncryptedPath)

	if len(result.OnlyInPlain) > 0 {
		sort.Strings(result.OnlyInPlain)
		fmt.Println("  Added (only in plain file):")
		for _, k := range result.OnlyInPlain {
			fmt.Printf("    + %s\n", k)
		}
	}

	if len(result.OnlyInEnc) > 0 {
		sort.Strings(result.OnlyInEnc)
		fmt.Println("  Removed (only in vault):")
		for _, k := range result.OnlyInEnc {
			fmt.Printf("    - %s\n", k)
		}
	}

	if len(result.Changed) > 0 {
		sort.Strings(result.Changed)
		fmt.Println("  Modified:")
		for _, k := range result.Changed {
			fmt.Printf("    ~ %s\n", k)
		}
	}

	// Exit with code 1 when diff is detected so it can be used in scripts.
	os.Exit(1)
	return nil
}
