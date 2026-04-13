package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runTemplate(args []string) {
	fs := flag.NewFlagSet("template", flag.ExitOnError)
	output := fs.String("output", "", "Destination path for the generated template (default: .env.template next to .env)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault template [--output <path>]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Generate a .env.template file from the encrypted vault.")
		fmt.Fprintln(os.Stderr, "All keys are preserved; values are replaced with empty strings.")
		fmt.Fprintln(os.Stderr)
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	dst, err := vault.GenerateTemplate(cfg, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("template written to %s\n", dst)
}
