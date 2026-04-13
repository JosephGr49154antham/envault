package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runSort(args []string) {
	fs := flag.NewFlagSet("sort", flag.ExitOnError)
	src := fs.String("src", "", "source .env file to sort (default: vault plain file)")
	dst := fs.String("dst", "", "destination file (default: in-place)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault sort [--src <file>] [--dst <file>]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Sort the keys in a .env file alphabetically.")
		fmt.Fprintln(os.Stderr, "Comments and blank lines are preserved at the top of the file.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	if err := vault.Sort(cfg, *src, *dst); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	effectiveDst := *dst
	if effectiveDst == "" {
		effectiveDst = *src
		if effectiveDst == "" {
			effectiveDst = cfg.PlainFile
		}
	}
	fmt.Printf("sorted: %s\n", effectiveDst)
}
