// Package vault manages the lifecycle of an envault project vault.
//
// A vault is a directory (default: .envault/) that lives alongside a project's
// source code and contains:
//
//   - recipients.txt  — the list of age public keys that are authorised to
//     decrypt the environment file.
//   - .env.age        — the age-encrypted copy of the project's .env file.
//
// Typical workflow:
//
//  1. Run Init to create the vault directory structure.
//  2. Use the recipients package (or the CLI) to add team-member public keys
//     to recipients.txt.
//  3. Call Push to encrypt the local .env and write .env.age.
//  4. Commit .envault/ (excluding any plaintext secrets) to version control.
//  5. Team members call Pull to decrypt .env.age into their local .env.
package vault
