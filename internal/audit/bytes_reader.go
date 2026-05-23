package audit

import "bytes"

// newBytesReader wraps a byte slice in a *bytes.Reader so it satisfies io.Reader.
func newBytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
