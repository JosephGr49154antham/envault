package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runFmt(args []string) {
	fs := flag.NewFlagSet("fmt", flag.ExitOnError)
	dst := fs.String("dst", "", "output path (default: overwrite source)")
	trim := fs.Bool("trim", true, "trim whitespace from values")
	quote := fs.Bool("quote", false, "quote values that contain spaces")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: envault fmt [flags] <src>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Normalise formatting of a .env file.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	cfg := vault.DefaultConfig()

	opts := vault.FmtOptions{
		Dst:         *dst,
		TrimValues:  *trim,
		QuoteValues: *quote,
	}

	if err := vault.Fmt(cfg, src, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := src
	if *dst != "" {
		out = *dst
	}
	fmt.Printf("formatted → %s\n", out)
}
