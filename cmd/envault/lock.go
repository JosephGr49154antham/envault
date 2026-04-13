package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/roamingthings/envault/internal/vault"
)

func runLock(args []string) {
	fs := flag.NewFlagSet("lock", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault lock")
		fmt.Fprintln(os.Stderr, "  Acquires an exclusive lock on the vault to prevent concurrent modifications.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()
	if err := vault.Lock(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Vault locked.")
}

func runUnlock(args []string) {
	fs := flag.NewFlagSet("unlock", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault unlock")
		fmt.Fprintln(os.Stderr, "  Releases the vault lock.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()
	if err := vault.Unlock(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Vault unlocked.")
}

func runLockStatus(args []string) {
	fs := flag.NewFlagSet("lock-status", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: envault lock-status")
		fmt.Fprintln(os.Stderr, "  Prints the current lock status of the vault.")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	cfg := vault.DefaultConfig()
	if !vault.IsInitialised(cfg) {
		fmt.Fprintln(os.Stderr, "error: vault is not initialised; run 'envault init' first")
		os.Exit(1)
	}
	fmt.Println(vault.LockStatus(cfg))
}
