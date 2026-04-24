package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

const resolveUsage = `Usage: envault resolve [options]

Expand variable references (${VAR} / $VAR) inside a .env file using values
defined earlier in the same file and from the current process environment.

Options:
  -src string   Source .env file (default: vault plain file)
  -dst string   Output file (default: overwrite source)
  -strict       Error if any reference cannot be resolved
`

func runResolve(args []string) error {
	fs := flag.NewFlagSet("resolve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	src := fs.String("src", "", "source .env file")
	dst := fs.String("dst", "", "output file")
	strict := fs.Bool("strict", false, "fail on unresolvable references")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig()

	opts := vault.ResolveOptions{
		Src:    *src,
		Dst:    *dst,
		Strict: *strict,
	}

	if err := vault.Resolve(cfg, opts); err != nil {
		return err
	}

	out := opts.Dst
	if out == "" {
		out = opts.Src
	}
	if out == "" {
		out = cfg.PlainFile
	}

	fmt.Fprintf(os.Stdout, "resolved: %s\n", out)
	return nil
}
