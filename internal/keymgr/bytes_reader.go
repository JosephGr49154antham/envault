package keymgr

import "bytes"

// bytesReader wraps a byte slice in a bytes.Reader for use with age.ParseIdentities.
func bytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
