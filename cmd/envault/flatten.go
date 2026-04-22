package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runFlatten(args []string) {
	fs := flag.NewFlagSet("flatten", flag.ExitOnError)
	prefix := fs.String("prefix", "", "prefix to prepend to every key")
	uppercase := fs.Bool("uppercase", false, "convert all keys to upper-case")
	dst := fs.String("dst", "", "output file path (default: <src>.flat<ext>)")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault flatten [flags] <src>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Normalise a .env file by optionally uppercasing keys and/or adding a prefix.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	cfg := vault.DefaultConfig()

	opts := vault.FlattenOptions{
		Prefix:    *prefix,
		Uppercase: *uppercase,
		Dst:       *dst,
		Overwrite: *overwrite,
	}

	out, err := vault.Flatten(cfg, src, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("flattened: %s\n", out)
}
