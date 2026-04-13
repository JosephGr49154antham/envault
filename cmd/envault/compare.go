package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/yourorg/envault/internal/vault"
)

func runCompare(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault compare <fileA> <fileB>")
		fmt.Fprintln(os.Stderr, "\nCompare two plain .env files and show their differences.")
		fs.PrintDefaults()
	}
	fs.Parse(args) //nolint:errcheck

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	fileA := fs.Arg(0)
	fileB := fs.Arg(1)

	cfg := vault.DefaultConfig()
	res, err := vault.Compare(cfg, fileA, fileB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !res.HasDifferences() {
		fmt.Println("✓ Files are identical.")
		return
	}

	if len(res.OnlyInA) > 0 {
		fmt.Printf("\nOnly in %s:\n", fileA)
		for _, k := range res.OnlyInA {
			fmt.Printf("  - %s\n", k)
		}
	}

	if len(res.OnlyInB) > 0 {
		fmt.Printf("\nOnly in %s:\n", fileB)
		for _, k := range res.OnlyInB {
			fmt.Printf("  + %s\n", k)
		}
	}

	if len(res.Changed) > 0 {
		fmt.Println("\nChanged values:")
		keys := make([]string, 0, len(res.Changed))
		for k := range res.Changed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			pair := res.Changed[k]
			fmt.Printf("  ~ %s: %q → %q\n", k, pair[0], pair[1])
		}
	}

	os.Exit(1) // non-zero exit when differences found, useful for CI
}
