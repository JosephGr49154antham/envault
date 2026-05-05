package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/envault/internal/vault"
)

func runUnset(args []string) error {
	fs := flag.NewFlagSet("unset", flag.ContinueOnError)
	src := fs.String("src", ".env", "source env file")
	dst := fs.String("dst", "", "output file (default: overwrite src)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	keys := fs.Args()
	if len(keys) == 0 {
		return fmt.Errorf("usage: envault unset [--src .env] KEY [KEY ...]")
	}

	cfg := vault.DefaultConfig()

	res, err := vault.Unset(cfg, vault.UnsetOptions{
		Src:  *src,
		Dst:  *dst,
		Keys: keys,
	})
	if err != nil {
		return err
	}

	for _, k := range res.Removed {
		fmt.Fprintf(os.Stdout, "removed: %s\n", k)
	}
	for _, k := range res.Missing {
		fmt.Fprintf(os.Stderr, "warning: key not found: %s\n", k)
	}
	return nil
}
