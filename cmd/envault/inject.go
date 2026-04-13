package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runInject(args []string) error {
	fs := flag.NewFlagSet("inject", flag.ExitOnError)
	src := fs.String("src", "", "encrypted env file to inject (default: vault encrypted file)")
	overwrite := fs.Bool("overwrite", false, "overwrite existing environment variables")
	dryRun := fs.Bool("dry-run", false, "print variables that would be injected without setting them")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault inject [flags] [-- command [args...]]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig()
	opts := vault.InjectOptions{
		Src:       *src,
		Overwrite: *overwrite,
		DryRun:    *dryRun,
	}

	result, err := vault.Inject(cfg, opts)
	if err != nil {
		return err
	}

	if *dryRun {
		fmt.Println("Variables that would be injected:")
		for _, k := range result.Set {
			fmt.Printf("  + %s\n", k)
		}
		for _, k := range result.Skipped {
			fmt.Printf("  ~ %s (skipped, already set)\n", k)
		}
		return nil
	}

	for _, k := range result.Set {
		fmt.Printf("injected: %s\n", k)
	}
	for _, k := range result.Skipped {
		fmt.Printf("skipped:  %s (already set)\n", k)
	}

	// If a sub-command was provided after --, exec it.
	remaining := fs.Args()
	if len(remaining) > 0 {
		path, err := exec.LookPath(remaining[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", remaining[0])
		}
		return syscall.Exec(path, remaining, os.Environ())
	}
	return nil
}
