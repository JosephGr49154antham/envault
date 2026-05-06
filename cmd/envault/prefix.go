package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runPrefix(args []string) {
	fs := flag.NewFlagSet("prefix", flag.ExitOnError)
	dst := fs.String("dst", "", "destination file (default: overwrite src)")
	prefix := fs.String("prefix", "", "prefix string to add or remove (required)")
	remove := fs.Bool("remove", false, "remove the prefix instead of adding it")
	overwrite := fs.Bool("overwrite", false, "allow overwriting an existing destination file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: envault prefix [flags] <src>\n\n")
		fmt.Fprintf(os.Stderr, "Add or remove a prefix from all keys in a .env file.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if *prefix == "" {
		fmt.Fprintln(os.Stderr, "error: --prefix is required")
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	if src == "" {
		fmt.Fprintln(os.Stderr, "error: src file argument is required")
		fs.Usage()
		os.Exit(1)
	}

	cfg := vault.DefaultConfig()
	err := vault.Prefix(cfg, vault.PrefixOptions{
		Src:       src,
		Dst:       *dst,
		Prefix:    *prefix,
		Remove:    *remove,
		Overwrite: *overwrite,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	action := "added to"
	if *remove {
		action = "removed from"
	}
	dstPath := src
	if *dst != "" {
		dstPath = *dst
	}
	fmt.Printf("prefix %q %s keys in %s\n", *prefix, action, dstPath)
}
