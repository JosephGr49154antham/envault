package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	values := fs.Bool("values", false, "search inside values as well as keys")
	noCase := fs.Bool("ignore-case", false, "case-insensitive matching")
	_ = fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "usage: envault search [--values] [--ignore-case] <pattern> [file...]")
		os.Exit(1)
	}

	pattern := remaining[0]
	files := remaining[1:]
	if len(files) == 0 {
		files = []string{".env"}
	}

	cfg := vault.DefaultConfig()
	opts := vault.SearchOptions{
		Pattern:      pattern,
		SearchValues: *values,
		CaseSensitive: !*noCase,
	}

	results, err := vault.Search(cfg, opts, files...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("no matches found")
		return
	}

	for _, r := range results {
		fmt.Printf("%s:%d\t%s=%s\n", r.File, r.Line, r.Key, r.Value)
	}
}
