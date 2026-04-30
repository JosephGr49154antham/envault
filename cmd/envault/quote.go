package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runQuote(args []string) {
	fs := flag.NewFlagSet("quote", flag.ExitOnError)
	dst := fs.String("dst", "", "destination file (default: overwrite src)")
	overwrite := fs.Bool("overwrite", false, "allow overwriting an existing destination file")
	force := fs.Bool("force", false, "re-quote values that are already quoted")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: envault quote [flags] <src>\n\n")
		fmt.Fprintf(os.Stderr, "Wrap every value in a .env file with double quotes.\n\n")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}
	src := fs.Arg(0)

	cfg := vault.DefaultConfig()
	opts := vault.QuoteOptions{
		Src:       src,
		Dst:       *dst,
		Overwrite: *overwrite,
		Force:     *force,
	}

	if err := vault.Quote(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := opts.Dst
	if out == "" {
		out = src
	}
	fmt.Printf("quoted: %s\n", out)
}
