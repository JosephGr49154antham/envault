package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nicholasgasior/envault/internal/vault"
)

// runAudit prints the vault audit log to stdout.
func runAudit(args []string) error {
	cfg := vault.DefaultConfig()

	if !vault.IsInitialised(cfg) {
		return fmt.Errorf("vault not initialised — run `envault init` first")
	}

	log, err := vault.LoadAuditLog(cfg)
	if err != nil {
		return fmt.Errorf("load audit log: %w", err)
	}

	if len(log.Events) == 0 {
		fmt.Println("No audit events recorded yet.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tOPERATION\tUSER\tDETAILS")
	for _, e := range log.Events {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.Operation,
			e.User,
			e.Details,
		)
	}
	return w.Flush()
}
