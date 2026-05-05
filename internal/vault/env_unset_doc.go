// Package vault – env_unset.go
//
// # Unset
//
// Unset removes one or more keys from a plain-text .env file.
// By default the source file is modified in place; supply a Dst path
// to write the result to a separate file and leave the original intact.
//
// # Behaviour
//
//   - Lines whose key matches one of the requested keys are dropped.
//   - Comment lines and blank lines are always preserved.
//   - Keys that do not exist in the file are reported in UnsetResult.Missing
//     rather than causing an error, so callers can decide how to handle them.
//
// # Example
//
//	res, err := vault.Unset(cfg, vault.UnsetOptions{
//	    Src:  ".env",
//	    Keys: []string{"SECRET_TOKEN", "OLD_KEY"},
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("removed: %v, not found: %v\n", res.Removed, res.Missing)
package vault
