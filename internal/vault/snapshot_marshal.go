package vault

import "encoding/json"

// marshalSnapshot serialises a Snapshot to JSON bytes.
// Extracted as a helper so both snapshot.go and tag_test.go can use it
// without duplicating the encoding logic.
func marshalSnapshot(s Snapshot) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
