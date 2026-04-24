package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/envault/internal/vault"
)

func runTrim(args []string) {
	fs := flag.NewFlagSet("trim", flag.ExitOnError)
	removeComments := fs.Bool("comments", false, "remove comment lines")
	removeBlanks := fs.Bool("blanks", false, "remove blank lines")
	trimValues := fs.Bool("values", false, "trim whitespace from keys and values")
	dst := fs.String("dst", "", "output file (default: overwrite src)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault trim [flags] <src>")
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

	if !*removeComments && !*removeBlanks && !*trimValues {
		fmt.Fprintln(os.Stderr, "trim: at least one of --comments, --blanks or --values must be set")
		os.Exit(1)
	}

	cfg := vault.DefaultConfig()
	err := vault.Trim(cfg, src, vault.TrimOptions{
		RemoveComments: *removeComments,
		RemoveBlanks:   *removeBlanks,
		TrimValues:     *trimValues,
		Dst:            *dst,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "trim: %v\n", err)
		os.Exit(1)
	}

	out := src
	if *dst != "" {
		out = *dst
	}
	fmt.Printf("trimmed → %s\n", out)
}
