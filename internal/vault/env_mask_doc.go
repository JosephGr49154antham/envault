/*
Package vault — Mask feature

# Mask

Mask reads a plain-text .env file and replaces sensitive values with a
fixed-length redaction string (default: "********"), producing a safe copy
that can be shared in logs, CI output, or documentation.

# Key selection

By default, Mask applies the same heuristic as Redact: any key whose name
contains a sensitive token (SECRET, PASSWORD, TOKEN, KEY, PRIVATE, CREDENTIAL,
API_KEY, etc.) has its value masked.

Pass an explicit Keys list to mask only those keys regardless of name.

# Usage

	envault mask [--src .env] [--dst .env.masked] [--keys KEY1,KEY2] [--char *] [--overwrite]

# Flags

	--src        Source .env file (default: .env)
	--dst        Destination file  (default: <src>.masked)
	--keys       Comma-separated list of keys to mask (overrides heuristic)
	--char       Mask character    (default: *)
	--overwrite  Replace destination if it already exists

# Differences from redact

Redact removes lines; Mask keeps all lines and replaces only the value
so consumers can still see which variables are defined.
*/
package vault
