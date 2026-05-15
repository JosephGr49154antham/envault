package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runFilter(args []string) {
	fs := flag.NewFlagSet("filter", flag.ExitOnError)
	src := fs.String("src", ".env", "source .env file")
	dst := fs.String("dst", "", "destination file (default: <src>.filtered)")
	pattern := fs.String("pattern", "", "key prefix to match (required)")
	negate := fs.Bool("negate", false, "exclude matching keys instead of keeping them")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault filter [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Filter .env keys by prefix pattern.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if *pattern == "" {
		fmt.Fprintln(os.Stderr, "error: --pattern is required")
		fs.Usage()
		os.Exit(1)
	}

	cfg := vault.DefaultConfig()
	err := vault.Filter(cfg, vault.FilterOptions{
		Src:       *src,
		Dst:       *dst,
		Pattern:   *pattern,
		Negate:    *negate,
		Overwrite: *overwrite,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := *dst
	if out == "" {
		out = "<default>"
	}
	fmt.Printf("filtered %q -> %s\n", *src, out)
}
