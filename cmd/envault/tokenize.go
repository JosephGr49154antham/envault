package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runTokenize(args []string) {
	fs := flag.NewFlagSet("tokenize", flag.ExitOnError)
	src := fs.String("src", ".env", "path to the .env file to tokenize")
	showErrors := fs.Bool("errors", false, "only print parse errors")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	result, err := vault.Tokenize(cfg, *src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *showErrors {
		if len(result.Errors) == 0 {
			fmt.Println("no parse errors")
			return
		}
		for _, e := range result.Errors {
			fmt.Println(e)
		}
		return
	}

	for _, tok := range result.Tokens {
		switch tok.Kind {
		case "key":
			fmt.Printf("[key]     line %3d  %-24s = %s\n", tok.Line, tok.Key, tok.Value)
		case "comment":
			fmt.Printf("[comment] line %3d  %s\n", tok.Line, tok.Raw)
		case "blank":
			fmt.Printf("[blank]   line %3d\n", tok.Line)
		case "invalid":
			fmt.Printf("[invalid] line %3d  %s\n", tok.Line, tok.Raw)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d parse error(s):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
		os.Exit(1)
	}
}
