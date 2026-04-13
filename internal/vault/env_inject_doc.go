/*
Package vault — Inject

# Overview

Inject decrypts an encrypted .env.age file and loads the key=value pairs
directly into the current process environment via os.Setenv.

This is useful when you want to run a sub-command with the secrets already
present in the environment without writing a plaintext file to disk.

# Usage

	result, err := vault.Inject(cfg, vault.InjectOptions{
		Src:       ".env.age",   // defaults to cfg.EncryptedFile when empty
		Overwrite: false,        // skip keys that already exist in the environment
		DryRun:    false,        // set true to preview without mutating the environment
	})

# Result fields

  - Set     — keys that were successfully injected.
  - Skipped — keys that were skipped because they already existed and
    Overwrite was false.

# Security note

Injected values live only in the memory of the current process and its
children. No plaintext file is written to disk.
*/
package vault
