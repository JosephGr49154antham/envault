package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/yourorg/envault/internal/vault"
)

func runMerge(args []string) {
	fs := flag.NewFlagSet("merge", flag.ExitOnError)
	dst := fs.String("dst", "", "output path (defaults to base path for in-place merge)")
	_ = fs.Parse(args)

	if fs.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: envault merge <base> <src> [--dst <output>]")
		os.Exit(1)
	}

	base := fs.Arg(0)
	src := fs.Arg(1)

	cfg := vault.DefaultConfig()

	res, err := vault.Merge(cfg, base, src, *dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge: %v\n", err)
		os.Exit(1)
	}

	outPath := *dst
	if outPath == "" {
		outPath = base + " (in-place)"
	}
	fmt.Printf("Merged into %s\n", outPath)

	if len(res.Added) > 0 {
		sort.Strings(res.Added)
		fmt.Printf("  + added   (%d): %v\n", len(res.Added), res.Added)
	}
	if len(res.Updated) > 0 {
		sort.Strings(res.Updated)
		fmt.Printf("  ~ updated (%d): %v\n", len(res.Updated), res.Updated)
	}
	if len(res.Kept) > 0 {
		fmt.Printf("  = kept    (%d) unchanged\n", len(res.Kept))
	}
}
