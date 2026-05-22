//go:build ignore
// +build ignore

// This file exists only to document the fmt import needed by history_test.go.
// The test file uses fmt.Sprintf via makeResults; in real builds the test
// package imports fmt directly.
package history_test

import "fmt"

var _ = fmt.Sprintf
