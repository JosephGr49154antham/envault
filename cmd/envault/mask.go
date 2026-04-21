package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envault/internal/vault"
)

func runMask(args []string) {
	fs := flag.NewFlagSet("mask", flag.ExitOnError)
	src := fs.String("src", ".env", "source .env file")
	dst := fs.String("dst", "", "destination file (default: <src>.masked)")
	keys := fs.String("keys", "", "comma-separated list of keys to mask")
	char := fs.String("char", "*", "character used to build the mask string")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it already exists")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	var keyList []string
	if *keys != "" {
		for _, k := range strings.Split(*keys, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				keyList = append(keyList, k)
			}
		}
	}

	opts := vault.MaskOptions{
		Src:       *src,
		Dst:       *dst,
		Keys:      keyList,
		MaskChar:  *char,
		Overwrite: *overwrite,
	}

	if err := vault.Mask(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out := opts.Dst
	if out == "" {
		out = *src + ".masked"
	}
	fmt.Printf("masked env written to %s\n", out)
}
