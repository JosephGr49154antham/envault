// Package vault provides the core operations for envault.
//
// # Dedup
//
// Dedup removes duplicate key definitions from a .env file, keeping either
// the first or the last occurrence of each key.
//
// By default the first occurrence is kept so that the file's original
// precedence order is preserved.  Pass KeepLast: true to keep the last
// occurrence instead — useful when you want later definitions to win, which
// mirrors the behaviour of most shell dotenv loaders.
//
// Comments and blank lines that appear between key definitions are preserved
// in the output.  A comment that immediately precedes a duplicate key that is
// being removed is also removed so that orphaned comments are not left behind.
//
// # Usage
//
//	err := vault.Dedup(vault.DedupOptions{
//		Cfg:      cfg,
//		Src:      ".env",
//		Dst:      ".env",       // omit to derive from Src
//		KeepLast: false,
//		Overwrite: true,
//	})
//
// # Example
//
// Given the following input:
//
//	# database
//	DB_HOST=localhost
//	DB_PORT=5432
//	DB_HOST=db.prod.example.com   # duplicate
//	APP_ENV=production
//
// Running Dedup with KeepLast: false produces:
//
//	# database
//	DB_HOST=localhost
//	DB_PORT=5432
//	APP_ENV=production
//
// Running Dedup with KeepLast: true produces:
//
//	DB_PORT=5432
//	DB_HOST=db.prod.example.com
//	APP_ENV=production
package vault
