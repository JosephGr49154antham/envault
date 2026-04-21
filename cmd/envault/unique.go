package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runUnique(args []string) {
	fs := flag.NewFlagSet("unique", flag.ExitOnError)
	src := fs.String("src", ".env", "source env file")
	dst := fs.String("dst", "", "output file (default: overwrite src)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault unique [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Remove duplicate keys from an env file, keeping the last occurrence.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	res, err := vault.Unique(cfg, *src, *dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(res.Removed) ==	fmt.Printf✔ Notreturn
	}
 Removed %d duplicate key %s\n", len(res.Removed), res.OutputPath)
	for _, k := range res.Removed {
		fmt.Printf("  - %s (earlier occurrence removed)\n", k)
	}
}
