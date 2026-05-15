/*
Package vault — Wrap

# Overview

Wrap folds long environment-variable values so that no line in the resulting
file exceeds a configurable column width (default 80).  Shell-style backslash
continuations are inserted between chunks, making the file easier to read in
terminals and code-review tools.

# Behaviour

  - Comment lines (# …) and blank lines are passed through unchanged.
  - KEY=VALUE lines whose total length is already within the limit are
    left as-is.
  - When a value must be split, subsequent continuation lines are indented
    to align with the first character of the value.
  - The destination file defaults to the source file (in-place rewrite).
  - Pass Overwrite: true to allow replacing an existing destination that
    differs from the source.

# Example

	err := vault.Wrap(cfg, vault.WrapOptions{
		Src:       ".env",
		Width:     72,
		Overwrite: true,
	})
*/
package vault
