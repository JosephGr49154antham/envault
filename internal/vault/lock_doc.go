// Package vault — lock.go
//
// # Vault Locking
//
// The lock feature prevents concurrent modifications to the encrypted vault
// by writing a lightweight lock file (.envault.lock) into the vault directory.
//
// ## Operations
//
//   - Lock(cfg)    — acquires the lock; fails if already locked
//   - Unlock(cfg)  — releases the lock; fails if not locked
//   - IsLocked(cfg) — returns true when a lock file is present
//   - LockStatus(cfg) — returns a human-readable description
//
// ## Lock file format
//
// The lock file is a single line of plain text:
//
//	user=alice machine=laptop locked_at=2024-01-15T10:00:00Z
//
// This makes it easy to inspect with cat/type without needing envault.
//
// ## Intended workflow
//
//	envault lock
//	# ... perform sensitive operations ...
//	envault unlock
//
// Push and pull automatically check for an existing lock before proceeding.
package vault
