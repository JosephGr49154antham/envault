package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envault/internal/vault"
)

func runPluck(args []string) error {
	fs := flag.NewFlagSet("pluck", flag.ContinueOnError)
	src := fs.String("src", "", "source .env file (default: plain file from vault config)")
	raw := fs.Bool("raw", false, "print bare values without KEY= prefix")

	if err := fs.Parse(args); err != nil {
		return err
	}

	keys := fs.Args()
	if len(keys) == 0 {
		return fmt.Errorf("usage: envault pluck [--src FILE] [--raw] KEY [KEY...]")
	}

	cfg := vault.DefaultConfig()
	opts := vault.PluckOptions{
		Src:  *src,
		Keys: keys,
		Raw:  *raw,
	}

	results, err := vault.Pluck(cfg, opts)
	if err != nil {
		return err
	}

	for _, line := range results {
		fmt.Fprintln(os.Stdout, line)
	}

	if !*raw {
		fmt.Fprintf(os.Stderr, "plucked %d key(s): %s\n",
			len(results), strings.Join(keys, ", "))
	}
	return nil
}
