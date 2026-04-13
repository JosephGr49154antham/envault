package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	overwrite := fs.Bool("overwrite", false, "overwrite existing encrypted file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: envault import [flags] <src-env-file>\n\n")
		fmt.Fprintf(os.Stderr, "Encrypt a plain .env file from <src-env-file> and store it in the vault.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	cfg := vault.DefaultConfig()

	if err := vault.Import(cfg, src, *overwrite); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Imported and encrypted %s → %s\n", src, cfg.EncryptedFile)
}
