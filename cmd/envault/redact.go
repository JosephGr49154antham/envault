package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rodrwan/envault/internal/vault"
)

func runRedact(args []string) {
	fs := flag.NewFlagSet("redact", flag.ExitOnError)
	src := fs.String("src", ".env", "source .env file to redact")
	dst := fs.String("dst", "", "destination file (default: <src>.redacted)")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	result, err := vault.Redact(cfg, *src, *dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath := *dst
	if outPath == "" {
		outPath = *src + ".redacted"
	}

	fmt.Printf("Redacted %d / %d keys(result.Redacted), result.Total, outPath)
	if len(result.Redacted) > 0 .Printf("  Redacted keys: %s\n", strings.Join(result.Redacted, ", "))
	}
}
