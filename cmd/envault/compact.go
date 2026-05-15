package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runCompact(args []string) {
	fs := flag.NewFlagSet("compact", flag.ExitOnError)
	dst := fs.String("dst", "", "destination file (defaults to src, in-place)")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault compact [flags] <src>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Remove blank lines and comments from an env file.")
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

	err := vault.Compact(cfg, vault.CompactOptions{
		Src:       src,
		Dst:       *dst,
		Overwrite: *overwrite,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := *dst
	if out == "" {
		out = src
	}
	fmt.Printf("compacted: %s\n", out)
}
