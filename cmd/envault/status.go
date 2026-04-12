package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nicholasgasior/envault/internal/vault"
)

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault status")
		fmt.Fprintln(os.Stderr, "\nShow sync status of encrypted env files in the vault.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := vault.DefaultConfig()

	statuses, err := vault.Status(cfg)
	if err != nil {
		return err
	}

	if len(statuses) == 0 {
		fmt.Println("No encrypted files found in vault.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tENCRYPTED\tSTATUS")

	for _, s := range statuses {
		status := statusLabel(s)
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.PlainFile, s.EncryptedFile, status)
	}

	return w.Flush()
}

func statusLabel(s vault.FileStatus) string {
	if !s.PlainExists {
		return "missing plain file"
	}
	if !s.EncExists {
		return "not encrypted"
	}
	if s.Stale {
		return "STALE (push needed)"
	}
	if s.InSync {
		return "ok"
	}
	return "unknown"
}
