package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/envault/internal/vault"
)

func runGrep(args []string) {
	fs := flag.NewFlagSet("grep", flag.ExitOnError)
	values := fs.Bool("values", false, "also search inside values")
	noCase := fs.Bool("i", false, "case-insensitive matching (default: true, use -i=false to disable)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault grep [flags] <pattern> [file ...]")
		fmt.Fprintln(os.Stderr, "\nSearch .env files for keys (and optionally values) matching a regex pattern.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	pattern := fs.Arg(0)
	files := fs.Args()[1:]
	if len(files) == 0 {
		files = []string{".env"}
	}

	cfg := vault.DefaultConfig()
	opts := vault.GrepOptions{
		Pattern:      pattern,
		SearchValues: *values,
		CaseSensitive: !*noCase,
	}

	results, err := vault.Grep(cfg, opts, files...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 .Println found.")
	
	 := range results {
	if len(files 1 {
	("%s:%d: %s\n", r.File, r.LineNum, r.Line)
		} else {
			fmt.Printf("%d: %s\n", r.LineNum, r.Line)
		}
	}
}
