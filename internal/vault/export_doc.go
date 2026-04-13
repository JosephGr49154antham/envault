// Package vault — Export feature
//
// # Overview
//
// Export decrypts the vault's encrypted env file and writes the plaintext to a
// local file that is intentionally kept outside the vault directory (and
// therefore outside version control).
//
// # Typical usage
//
//	// Export to an auto-named file next to the project root:
//	path, err := vault.Export(cfg, vault.ExportOptions{})
//
//	// Export to a specific path, allowing overwrite:
//	path, err := vault.Export(cfg, vault.ExportOptions{
//	    OutputPath: "/tmp/staging.env",
//	    Overwrite:  true,
//	})
//
// # Security notes
//
//   - The caller's identity file (cfg.IdentityFile) must exist and correspond
//     to one of the recipients stored in cfg.RecipientsFile.
//   - The exported plaintext file has mode 0600 (set by crypto.DecryptFile).
//   - The auto-generated filename embeds a timestamp to avoid accidental
//     collisions but does NOT guarantee uniqueness under concurrent calls.
//   - Users should add the exported file pattern to .gitignore.
package vault
