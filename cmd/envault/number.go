package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runNumber(args []string) {
	fs := flag.NewFlagSet("number", flag.ExitOnError)
	dst := fs.String("dst", "", "destination file (default: <src>.numbered)")
	onlyKV := fs.Bool("kv", false, "number only key=value lines; leave comments and blanks unnumbered")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: envault number [flags] <src>\n\n")
		fmt.Fprintf(os.Stderr, "Prepend line numbers to every line of a .env file.\n\n")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	cfg := vault.DefaultConfig()

	err := vault.Number(cfg, vault.NumberOptions{
		Src:    src,
		Dst:    *dst,
		OnlyKV: *onlyKV,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := *dst
	if out == "" {
		out = src + ".numbered"
	}
	fmt.Printf("numbered env written to %s\n", out)
}
