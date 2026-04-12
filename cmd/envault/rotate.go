package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nicholasgasior/envault/internal/keymgr"
	"github.com/nicholasgasior/envault/internal/vault"
)

// runRotate handles the `envault rotate` sub-command.
// It backs up the current encrypted file, then re-encrypts it for the
// current recipients list using the caller's identity for decryption.
func runRotate(args []string) {
	fs := flag.NewFlagSet("rotate", flag.ExitOnError)
	identityPath := fs.String("identity", keymgr.DefaultIdentityPath(), "path to age identity file")
	backupDir := fs.String("backup-dir", "", "directory for encrypted-file backups (default: <vault>/.envault/backups)")
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()

	if !vault.IsInitialised(cfg) {
		fmt.Fprintln(os.Stderr, "error: vault not initialised. Run 'envault init' first.")
		os.Exit(1)
	}

	_, err := keymgr.LoadIdentity(*identityPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading identity %s: %v\n", *identityPath, err)
		os.Exit(1)
	}

	opts := vault.RotateOptions{BackupDir: *backupDir}
	if err := vault.Rotate(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "rotate failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Vault rotated successfully.")
	if *backupDir != "" {
		fmt.Printf("  Backup stored in: %s\n", *backupDir)
	} else {
		fmt.Printf("  Backup stored in: %s/backups\n", cfg.VaultDir)
	}
}
