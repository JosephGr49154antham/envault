package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runWrap(args []string) {
	fs := flag.NewFlagSet("wrap", flag.ExitOnError)
	src := fs.String("src", ".env", "source env file")
	dst := fs.String("dst", "", "destination file (defaults to src)")
	width := fs.Int("width", 80, "maximum line width")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	err := vault.Wrap(cfg, vault.WrapOptions{
		Src:       *src,
		Dst:       *dst,
		Width:     *width,
		Overwrite: *overwrite,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	destination := *dst
	if destination == "" {
		destination = *src
	}
	fmt.Printf("wrapped %q → %q (width %d)\n", *src, destination, *width)
}
