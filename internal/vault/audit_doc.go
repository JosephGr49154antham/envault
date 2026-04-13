// Package vault provides the core vault operations for envault.
//
// # Audit Log
//
// The audit log records every mutating vault operation (push, pull, rekey,
// rotate, add-recipient) to a JSON file stored at
//
//	<vault-dir>/audit.json
//
// Each entry captures:
//   - timestamp  – UTC time of the operation
//   - operation  – name of the operation (e.g. "push", "rekey")
//   - user       – optional identifier of the actor (e.g. git user.email)
//   - details    – optional free-form description
//
// The log is append-only from the perspective of the public API; entries are
// never removed or modified by envault itself.  Teams may choose to commit
// audit.json alongside the encrypted file for a lightweight change history.
//
// Example audit.json:
//
//	{
//	  "events": [
//	    {
//	      "timestamp": "2024-01-15T10:30:00Z",
//	      "operation": "push",
//	      "user": "alice@example.com",
//	      "details": "initial push"
//	    }
//	  ]
//	}
package vault
