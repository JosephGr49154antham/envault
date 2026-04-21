// Package vault — env_patch.go
//
// # Patch
//
// Patch applies a declarative set of mutations to an .env file:
//
//   - Upserts: add a key if absent, or update its value if already present.
//   - Deletions: remove a key (and its value line) entirely.
//
// Comments and blank lines are preserved in their original positions so that
// hand-written documentation inside .env files is not lost.
//
// # Usage
//
//	err := vault.Patch(cfg, vault.PatchOptions{
//		Upserts: map[string]string{
//			"DATABASE_URL": "postgres://localhost/prod",
//			"NEW_FEATURE":  "true",
//		},
//		Deletions: []string{"LEGACY_FLAG"},
//	})
//
// # Atomicity
//
// The output is written to Dst (defaulting to Src for an in-place edit) only
// after all mutations have been applied in memory, so a partial failure cannot
// leave a half-written file.
package vault
