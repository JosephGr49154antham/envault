package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/your-org/envault/internal/vault"
)

func runRenameValue(args []string) {
	fs := flag.NewFlagSet("rename-value", flag.ExitOnError)
	src := fs.String("src", ".env", "source env file")
	dst := fs.String("dst", "", "destination file (defaults to src)")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault rename-value [flags] <old-value> <new-value>")
		fmt.Fprintln(os.Stderr, "\nReplace all occurrences of a value in an env file.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	oldVal := fs.Arg(0)
	newVal := fs.Arg(1)

	cfg := vault.DefaultConfig()
	err := vault.RenameValue(cfg, vault.RenameValueOptions{
		Src:      *src,
		Dst:      *dst,
		OldValue: oldVal,
		NewValue: newVal,
		Overwrite: *overwrite,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dstPath := *dst
	if dstPath == "" {
		dstPath = *src
	}
	fmt.Printf("Replaced %q → %q in %s\n", oldVal, newVal, dstPath)
}
