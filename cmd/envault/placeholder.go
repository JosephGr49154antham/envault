package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runPlaceholder(args []string) {
	fs := flag.NewFlagSet("placeholder", flag.ExitOnError)
	src := fs.String("src", ".env", "source .env file")
	dst := fs.String("dst", "", "output file (default: <src>.placeholder)")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")
	valueFmt := fs.String("format", "<%s>", "printf format string for placeholder values (receives key name)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault placeholder [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Generate a .env file with all values replaced by safe placeholders.")
		fmt.Fprintln(os.Stderr, "Useful for committing a documented template without exposing secrets.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	out, err := vault.GeneratePlaceholder(cfg, vault.PlaceholderOptions{
		Src:       *src,
		Dst:       *dst,
		Overwrite: *overwrite,
		ValueFmt:  *valueFmt,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("placeholder written to %s\n", out)
}
