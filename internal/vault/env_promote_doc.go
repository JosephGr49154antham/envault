// Package vault — env_promote.go
//
// # Promote
//
// Promote copies environment variables from one .env file (source) into
// another (destination) without overwriting any keys that already exist in
// the destination. This is useful for safely propagating new configuration
// from a lower environment (e.g. staging) to a higher one (e.g. production).
//
// ## Behaviour
//
//   - Keys present in src but absent in dst are added to dst.
//   - Keys already present in dst are left untouched regardless of their
//     value in src.
//   - If dst does not exist it is created.
//   - The vault must be initialised before Promote is called.
//
// ## Return value
//
// Promote returns a PromoteResult that lists:
//   - KeysAdded  — keys copied from src into dst
//   - KeysKept   — keys that existed in dst and were not overwritten
//
// ## Example
//
//	res, err := vault.Promote(cfg, ".env.staging", ".env.production")
//	if err != nil { log.Fatal(err) }
//	fmt.Printf("%d key(s) added\n", len(res.KeysAdded))
package vault
