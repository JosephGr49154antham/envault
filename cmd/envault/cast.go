package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/envault/internal/vault"
)

func runCast(args []string) error {
	fs := flag.NewFlagSet("cast", flag.ContinueOnError)
	src := fs.String("src", "", "source env file (default: vault plain file)")
	dst := fs.String("dst", "", "output file (default: overwrite src)")
	keys := fs.String("keys", "", "comma-separated list of keys to cast (default: all)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig()

	var keyList []string
	if *keys != "" {
		for _, k := range strings.Split(*keys, ",") {
			if k = strings.TrimSpace(k); k != "" {
				keyList = append(keyList, k)
			}
		}
	}

	results, err := vault.Cast(cfg, vault.CastOptions{
		Src:  *src,
		Dst:  *dst,
		Keys: keyList,
	})
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stdout, "no values were cast")
		return nil
	}

	changed := 0
	for _, r := range results {
		if r.OK {
			fmt.Fprintf(os.Stdout, "  %-20s %s -> %s  (%s)\n", r.Key, r.Original, r.Cast, r.Type)
			changed++
		}
	}
	if changed == 0 {
		fmt.Fprintln(os.Stdout, "all values already canonical")
	} else {
		fmt.Fprintf(os.Stdout, "%d value(s) rewritten\n", changed)
	}
	return nil
}
