package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runRenameKey(cfg vault.Config, args []string) {
	fs := flag.NewFlagSet("rename-key", flag.ExitOnError)
	dst := fs.String("dst", "", "destination file (default: overwrite src)")
	force := fs.Bool("force", false, "overwrite destination if it already exists")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault rename-key [flags] <src> <old-key> <new-key>")
		fmt.Fprintln(os.Stderr, "\nRename a key inside an env file.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() != 3 {
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	oldKey := fs.Arg(1)
	newKey := fs.Arg(2)

	opts := vault.RenameKeyOptions{
		Src:    src,
		Dst:    *dst,
		OldKey: oldKey,
		NewKey: newKey,
		Force:  *force,
	}

	if err := vault.RenameKey(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := src
	if *dst != "" {
		out = *dst
	}
	fmt.Printf("renamed %q → %q in %s\n", oldKey, newKey, out)
}
