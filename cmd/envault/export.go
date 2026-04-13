package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/envault/envault/internal/vault"
)

const exportUsage = `Usage: envault export [flags]

Decrypt the vault's encrypted env file and write plaintext to a local file.
The output file is never added to git.

Flags:
  -output string   Destination path (default: auto-generated timestamped name)
  -overwrite       Overwrite the destination if it already exists
`

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	output := fs.String("output", "", "destination file path")
	overwrite := fs.Bool("overwrite", false, "overwrite existing destination")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig()

	opts := vault.ExportOptions{
		OutputPath: *output,
		Overwrite:  *overwrite,
	}

	path, err := vault.Export(cfg, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Printf("exported → %s\n", path)
	return nil
}
