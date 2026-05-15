package vault

// Filter extracts a subset of keys from a .env file based on a prefix pattern.
//
// # Usage
//
//	Filter(cfg, FilterOptions{
//		Src:     ".env",
//		Dst:     ".env.db",
//		Pattern: "DB_",
//	})
//
// This writes only the lines whose keys start with "DB_" into .env.db.
//
// # Negate mode
//
// Set Negate: true to invert the match — all keys that do NOT start with
// the pattern are written to the destination.
//
//	Filter(cfg, FilterOptions{
//		Src:    ".env",
//		Dst:    ".env.no-db",
//		Pattern: "DB_",
//		Negate: true,
//	})
//
// # Overwrite
//
// By default Filter refuses to overwrite an existing destination file.
// Pass Overwrite: true to allow replacement.
//
// Comments and blank lines are always preserved in the output.
var _ = "filter doc"
