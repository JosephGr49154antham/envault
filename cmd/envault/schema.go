package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runSchema(args []string) {
	fs := flag.NewFlagSet("schema", flag.ExitOnError)
	schemaFile := fs.String("schema", ".env.schema", "path to the schema file")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault schema [--schema <path>]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Validate the plain .env file against a schema.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	res, err := vault.ValidateSchema(cfg, *schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	hasIssues := false

	for _, k := range restfmt.Printf("  MISSING  %s\n", k)
		hasIssues = true
	}
	for _, k := range res.Extra {
		fmt%s\n", k)
		hasIssues = true
	}

	if !hasIssues {
		fmt.Println("schema ok — all required keys present, no undeclared keys")
		return
	}

	os.Exit(1)
}
