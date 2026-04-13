package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envault/internal/vault"
)

// runSplit implements the `envault split` sub-command.
//
// Usage:
//
//	envault split [flags] <src>
//	  -prefix DB_=services/db/.env -prefix REDIS_=services/redis/.env
//	  -remainder services/app/.env
func runSplit(args []string) {
	fs := flag.NewFlagSet("split", flag.ExitOnError)

	var rawPrefixes prefixFlag
	fs.Var(&rawPrefixes, "prefix", "PREFIX=dest mapping (repeatable); e.g. DB_=db.env")

	remainder := fs.String("remainder", "", "destination file for keys that match no prefix")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault split [flags] <src>")
		fs.PrintDefaults()
	}

	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}
	src := fs.Arg(0)

	prefixes := make(map[string]string, len(rawPrefixes))
	for _, p := range rawPrefixes {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			fmt.Fprintf(os.Stderr, "split: invalid -prefix value %q (want PREFIX=dest)\n", p)
			os.Exit(1)
		}
		prefixes[parts[0]] = parts[1]
	}

	cfg := vault.DefaultConfig()
	opts := vault.SplitOptions{
		Prefixes:  prefixes,
		Remainder: *remainder,
	}

	if err := vault.Split(cfg, src, opts); err != nil {
		fmt.Fprintf(os.Stderr, "split: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("split %s into %d file(s)\n", src, len(prefixes)+boolInt(*remainder != ""))
}

// prefixFlag is a repeatable string flag.
type prefixFlag []string

func (p *prefixFlag) String() string  { return strings.Join(*p, ", ") }
func (p *prefixFlag) Set(v string) error { *p = append(*p, v); return nil }

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
