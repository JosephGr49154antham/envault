// Package vault — Interpolate
//
// # Overview
//
// Interpolate expands ${VAR} and $VAR references that appear in .env values,
// using the variables defined within the same file.  An optional overlay file
// can supply additional or overriding values (useful for environment-specific
// overrides such as .env.local).
//
// # Behaviour
//
//   - Comment lines (# …) and blank lines are copied unchanged.
//   - Only the value side of each KEY=VALUE pair is expanded.
//   - If a referenced variable is not defined in the file (or overlay), the
//     tool falls back to the process environment before leaving the reference
//     as-is (or returning an error in strict mode).
//   - The output is written to a new file; the source is never modified.
//
// # Strict mode
//
// When InterpolateOptions.Strict is true, any reference to an undefined
// variable causes Interpolate to return an error immediately.
//
// # Example
//
//	envault interpolate --src .env --overlay .env.local --strict
package vault
