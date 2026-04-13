package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runPromote(args []string) {
	fs := flag.NewFlagSet("promote", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault promote <src> <dst>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Copy new keys from src env file into dst env file.")
		fmt.Fprintln(os.Stderr, "Keys already present in dst are never overwritten.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	src := fs.Arg(0)
	dst := fs.Arg(1)

	cfg := vault.DefaultConfig(".")
	res, err := vault.Promote(cfg, src, dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(res.KeysAdded) == 0 {
		fmt.Println("No new keys to promote — destination is already up to date.")
		return
	}

	fmt.Printf("Promoted %d key(s) from %s → %s\n", len(res.KeysAdded), srcfor _, k := range res.Keys %s\n", k)
	}
	if len(res.KeysKept) > 0 {
		fmt.Printf("Kept %d existing key(s) unchanged.\n", len(res.KeysKept))
	}
}
