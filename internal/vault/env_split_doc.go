/*
Package vault — Split

# Overview

Split partitions a single .env file into multiple output files based on key
prefixes. This is useful when a monolithic env file needs to be broken up by
service (e.g. DB_, REDIS_, APP_).

# Usage

	opts := vault.SplitOptions{
		Prefixes: map[string]string{
			"DB_":    "services/db/.env",
			"REDIS_": "services/redis/.env",
		},
		Remainder: "services/app/.env",
	}
	err := vault.Split(cfg, ".env", opts)

# Rules

  - Comment lines and blank lines are copied to every output file.
  - A key is matched against prefixes in iteration order; the first match wins.
  - Keys that match no prefix are written to Remainder (if set); otherwise
    they are silently dropped.
  - Output directories are created automatically.
  - The vault must be initialised before calling Split.
*/
package vault
