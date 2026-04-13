// Package vault provides the core operations for envault.
//
// # Merge
//
// Merge combines two .env files (a base and a source) into a single output.
// It is useful when pulling updates from a shared vault while preserving
// local-only keys.
//
// Rules:
//
//   - Keys present only in src  → added to output (Added)
//   - Keys present in both with different values → src value wins (Updated)
//   - Keys present in both with the same value   → kept unchanged (Kept)
//   - Keys present only in base → kept unchanged (not listed separately)
//
// The destination path defaults to the base path when left empty, performing
// an in-place merge.
//
// Example usage:
//
//	result, err := vault.Merge(cfg, ".env", ".env.new", "")
//	if err != nil { ... }
//	fmt.Printf("added %d, updated %d\n", len(result.Added), len(result.Updated))
package vault
