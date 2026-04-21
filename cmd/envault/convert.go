package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/envault/internal/vault"
)

func runConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	format := fs.String("format", "dotenv", "output format: dotenv, export, json, yaml")
	dst := fs.String("out", "", "destination file (default: derived from source)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: envault convert [flags] <src>")
	}

	src := fs.Arg(0)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source file not found: %s", src)
	}

	cfg := vault.DefaultConfig()

	out, err := vault.Convert(cfg, vault.ConvertOptions{
		Src:    src,
		Dst:    *dst,
		Format: vault.ConvertFormat(*format),
	})
	if err != nil {
		return err
	}

	fmt.Printf("converted %s → %s\n", src, out)
	return nil
}
