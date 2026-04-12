// Package main is the entry point for the envault CLI tool.
// It wires together the vault, crypto, keymgr, and recipients packages
// and exposes them as subcommands via a simple flag-based CLI.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/envault/internal/keymgr"
	"github.com/yourusername/envault/internal/recipients"
	"github.com/yourusername/envault/internal/vault"
)

const usage = `envault — encrypt and sync .env files using age encryption.

Usage:
  envault <command> [flags]

Commands:
  init          Initialise a new vault in the current directory
  keygen        Generate a new age key pair and save the identity file
  add-recipient Add a teammate's public key to the vault recipients list
  push          Encrypt .env and push to the vault store
  pull          Decrypt the stored vault file back to .env

Run "envault <command> -help" for command-specific flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "keygen":
		runKeygen(os.Args[2:])
	case "add-recipient":
		runAddRecipient(os.Args[2:])
	case "push":
		runPush(os.Args[2:])
	case "pull":
		runPull(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

// runInit initialises a new vault in the current directory.
func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	fs.Parse(args) //nolint:errcheck

	cfg := vault.DefaultConfig()
	if err := vault.Init(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: init failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Vault initialised successfully.")
	fmt.Printf("  Recipients : %s\n", cfg.RecipientsFile)
	fmt.Printf("  Vault store: %s\n", cfg.EncryptedFile)
}

// runKeygen generates a new age identity and prints the public key.
func runKeygen(args []string) {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	output := fs.String("o", keymgr.DefaultIdentityPath(), "path to write the identity file")
	fs.Parse(args) //nolint:errcheck

	id, err := keymgr.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: keygen failed: %v\n", err)
		os.Exit(1)
	}
	if err := keymgr.SaveIdentity(id, *output); err != nil {
		fmt.Fprintf(os.Stderr, "error: saving identity: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Identity saved to %s\n", *output)
	fmt.Printf("Public key : %s\n", id.Recipient().String())
	fmt.Println("Share the public key with your team so they can add you as a recipient.")
}

// runAddRecipient adds a public key to the vault recipients list.
func runAddRecipient(args []string) {
	fs := flag.NewFlagSet("add-recipient", flag.ExitOnError)
	fs.Parse(args) //nolint:errcheck

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: envault add-recipient <age-public-key>")
		os.Exit(1)
	}

	cfg := vault.DefaultConfig()
	pubKey := fs.Arg(0)
	if err := recipients.AddRecipient(cfg.RecipientsFile, pubKey); err != nil {
		fmt.Fprintf(os.Stderr, "error: add-recipient failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Recipient added: %s\n", pubKey)
}

// runPush encrypts the .env file and writes the ciphertext to the vault store.
func runPush(args []string) {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	envFile := fs.String("f", ".env", "path to the plaintext .env file")
	fs.Parse(args) //nolint:errcheck

	cfg := vault.DefaultConfig()
	identityPath := keymgr.DefaultIdentityPath()

	id, err := keymgr.LoadIdentity(identityPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: loading identity from %s: %v\n", identityPath, err)
		os.Exit(1)
	}

	if err := vault.Push(cfg, *envFile, id); err != nil {
		fmt.Fprintf(os.Stderr, "error: push failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Pushed %s → %s\n", *envFile, cfg.EncryptedFile)
}

// runPull decrypts the vault store back to a .env file.
func runPull(args []string) {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	envFile := fs.String("f", ".env", "path to write the decrypted .env file")
	fs.Parse(args) //nolint:errcheck

	cfg := vault.DefaultConfig()
	identityPath := keymgr.DefaultIdentityPath()

	id, err := keymgr.LoadIdentity(identityPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: loading identity from %s: %v\n", identityPath, err)
		os.Exit(1)
	}

	if err := vault.Pull(cfg, *envFile, id); err != nil {
		fmt.Fprintf(os.Stderr, "error: pull failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Pulled %s → %s\n", cfg.EncryptedFile, *envFile)
}
