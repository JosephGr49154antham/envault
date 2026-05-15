package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/vault"
)

func runJoin(args []string) {
	fs := flag.NewFlagSet("join", flag.ExitOnError)
	dst := fs.String("dst", "", "output file path (default: <vaultDir>/.env.joined)")
	overwrite := fs.Bool("overwrite", false, "overwrite destination if it exists")
	separator := fs.Bool("separator", false, "add a comment separator between each source file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: envault join [flags] <file1> [file2 ...]\n\n")
		fmt.Fprintf(os.Stderr, "Merge multiple .env files into a single output file.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	srcs := fs.Args()
	if len(srcs) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one source file is required")
		fs.Usage()
		os.Exit(1)
	}

	cfg := vault.DefaultConfig()
	opts := vault.JoinOptions{
		Dst:       *dst,
		Overwrite: *overwrite,
		Separator: *separator,
	}

	if err := vault.Join(cfg, srcs, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath := opts.Dst
	if outPath == "" {
		outPath = cfg.VaultDir + "/.env.joined"
	}
	fmt.Printf("joined %d file(s) → %s\n", len(srcs), outPath)
}
