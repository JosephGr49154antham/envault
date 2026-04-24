/*
Package vault – Trim

# Overview

Trim applies one or more clean-up transformations to a .env file:

  - RemoveComments – deletes lines that begin with '#'.
  - RemoveBlanks   – deletes lines that are empty or contain only whitespace.
  - TrimValues     – strips leading/trailing whitespace from both the key and
    the value side of every KEY=VALUE pair.

Any combination of the three flags may be used together.

# Destination

By default Trim writes the result back to the source file (in-place edit).
Set TrimOptions.Dst to write to a separate output path and leave the
original file untouched.

# Example

	err := vault.Trim(cfg, ".env", vault.TrimOptions{
		RemoveComments: true,
		RemoveBlanks:   true,
		TrimValues:     true,
	})
*/
package vault
