package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envault/internal/vault"
)

func runLower(args []string) error {
	fs := flag.NewFlagSet("lower", flag.ContinueOnError)
	src := fs.String("src", ".env", "source .env file")
	dst := fs.String("dst", "", "destination file (default: <src>.lower)")
	keys := fs.String("keys", "", "comma-separated list of keys to lowercase (default: all)")
	inPlace := fs.Bool("in-place", false, "overwrite source file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	destPath := *dst
	if destPath == "" && !*inPlace {
		destPath = *src + ".lower"
	}
	if *inPlace {
		destPath = *src
	}

	var keyList []string
	if *keys != "" {
		for _, k := range strings.Split(*keys, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				keyList = append(keyList, k)
			}
		}
	}

	if err := vault.Lower(*src, destPath, keyList, *inPlace); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Printf("lower: wrote %s\n", destPath)
	return nil
}
