// Package vault provides the core operations for the envault CLI tool.
//
// # Diff
//
// The Diff function compares a plain .env file against the encrypted vault
// file, reporting which keys have been added, removed, or changed without
// exposing their values.
//
// Usage:
//
//	result, err := vault.Diff(cfg, ".env")
//	if err != nil { ... }
//	if result.HasDiff() {
//	    // handle differences
//	}
//
// DiffResult fields:
//
//	OnlyInPlain  – keys present in the plain file but not in the vault
//	OnlyInEnc    – keys present in the vault but not in the plain file
//	Changed      – keys whose values differ between plain and vault
//	Unchanged    – keys that are identical in both files
//
// The diff operation requires a valid age identity to decrypt the vault file.
// The identity is loaded from the default path (~/.config/envault/identity.age).
package vault
