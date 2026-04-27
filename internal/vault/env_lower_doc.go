// Package vault provides the Lower operation for envault.
//
// # Lower
//
// Lower reads a .env file and converts the values of all (or selected) keys
// to lowercase, writing the result to a destination file.
//
// Comments and blank lines are preserved verbatim.
//
// # Usage
//
//	err := vault.Lower(src, dst, keys, inPlace)
//
// Parameters:
//   - src      – path to the source .env file
//   - dst      – path to write the lowercased output
//   - keys     – if non-nil/non-empty, only lowercase values for these keys
//   - inPlace  – if true, overwrite src (dst is ignored)
//
// # Example
//
//	// Lowercase all values
//	vault.Lower(".env", ".env.lower", nil, false)
//
//	// Lowercase only DB_PASSWORD in-place
//	vault.Lower(".env", "", []string{"DB_PASSWORD"}, true)
package vault
