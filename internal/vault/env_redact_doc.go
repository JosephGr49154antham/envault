// Package vault — Redact
//
// # Overview
//
// Redact scans a plain-text .env file and replaces the values of any keys
// that look sensitive with the placeholder "***REDACTED***". The original
// file is never modified; a new file is written to the destination path.
//
// # Sensitive key detection
//
// A key is considered sensitive when its name (case-insensitive) contains
// any of the following substrings:
//
//	secret  password  passwd  token  apikey  api_key
//	private credential auth    jwt    passphrase
//
// # Usage
//
//	result, err := vault.Redact(cfg, ".env", ".env.redacted")
//	if err != nil { ... }
//	fmt.Printf("Redacted %d / %d keys\n", len(result.Redacted), result.Total)
//
// # Default paths
//
// If src is empty it defaults to ".env".
// If dst is empty it defaults to "<src>.redacted" (e.g. ".env.redacted").
//
// # Notes
//
//   - Blank lines and comment lines (starting with #) are preserved as-is.
//   - Lines that do not contain "=" are passed through unchanged.
//   - The vault must be initialised before calling Redact.
package vault
