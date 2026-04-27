package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/envault/internal/vault"
)

func runEcho(args []string) {
	fs := flag.NewFlagSet("echo", flag.ExitOnError)
	src := fs.String("src", "", "source .env file (default: vault plain file)")
	export := fs.Bool("export", false, "prefix each line with 'export '")
	quote := fs.Bool("quote", false, "wrap values in double quotes")
	nullSep := fs.Bool("0", false, "separate output with null bytes instead of newlines")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault echo [flags] [KEY...]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Print env key=value pairs to stdout.")
		fmt.Fprintln(os.Stderr, "Optionally filter to specific keys.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	keys := fs.Args()

	cfg := vault.DefaultConfig()

	// Validate explicit keys look like env var names
	for _, k := range keys {
		if strings.ContainsAny(k, " =") {
			fmt.Fprintf(os.Stderr, "echo: invalid key %q\n", k)
			os.Exit(1)
		}
	}

	opts := vault.EchoOptions{
		Src:     *src,
		Keys:    keys,
		Export:  *export,
		Quote:   *quote,
		NullSep: *nullSep,
	}

	if err := vault.Echo(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "echo: %v\n", err)
		os.Exit(1)
	}
}
