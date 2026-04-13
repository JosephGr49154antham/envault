package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runRename(args []string) {
	fs := flag.NewFlagSet("rename", flag.ExitOnError)
	src := fs.String("f", ".env", "path to the .env file to modify")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault rename [flags] <OLD_KEY> <NEW_KEY>")
		fmt.Fprintln(os.Stderr, "\nRename an environment variable key within a .env file.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	oldKey := fs.Arg(0)
	newKey := fs.Arg(1)

	cfg := vault.DefaultConfig()

	if err := vault.Rename(cfg, *src, oldKey, newKey); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✔ Renamed %q → %q in %s\n", oldKey, newKey, *src)
}
