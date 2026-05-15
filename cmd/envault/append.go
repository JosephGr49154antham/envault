package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envault/internal/vault"
)

func runAppend(args []string) error {
	fs := flag.NewFlagSet("append", flag.ContinueOnError)
	keys := fs.String("keys", "", "comma-separated list of keys to append (default: all new keys)")
	noBlank := fs.Bool("no-blank", false, "skip blank lines from source")
	dryRun := fs.Bool("dry-run", false, "report keys that would be appended without modifying destination")
	dst := fs.String("dst", "", "destination .env file (default: .env)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: envault append [flags] <src>")
	}

	src := fs.Arg(0)
	dstPath := *dst
	if dstPath == "" {
		dstPath = ".env"
	}

	var keyList []string
	if *keys != "" {
		for _, k := range strings.Split(*keys, ",") {
			if k = strings.TrimSpace(k); k != "" {
				keyList = append(keyList, k)
			}
		}
	}

	cfg := vault.DefaultConfig()
	appended, err := vault.Append(cfg, vault.AppendOptions{
		Src:     src,
		Dst:     dstPath,
		Keys:    keyList,
		NoBlank: *noBlank,
		DryRun:  *dryRun,
	})
	if err != nil {
		return err
	}

	if len(appended) == 0 {
		fmt.Fprintln(os.Stdout, "no new keys to append")
		return nil
	}

	if *dryRun {
		fmt.Fprintf(os.Stdout, "would append %d key(s): %s\n", len(appended), strings.Join(appended, ", "))
	} else {
		fmt.Fprintf(os.Stdout, "appended %d key(s) to %s: %s\n", len(appended), dstPath, strings.Join(appended, ", "))
	}
	return nil
}
