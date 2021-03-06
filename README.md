# byteline


[![Go Report Card](https://goreportcard.com/badge/github.com/erinpentecost/byteline)](https://goreportcard.com/report/github.com/erinpentecost/byteline)
[![Travis CI](https://travis-ci.org/erinpentecost/byteline.svg?branch=master)](https://travis-ci.org/erinpentecost/byteline)
[![GoDoc](https://godoc.org/github.com/erinpentecost/byteline?status.svg)](https://godoc.org/github.com/erinpentecost/byteline)

Map byte offsets to line + column and back with a writer/reader middleware!

## Features

* Supports `/r`, `/n`, `/r/n`, `/n/r` newline declarations without getting confused by `/r/r` or `/n/n`.
* Lines and offsets are reported correctly even if a `rune`'s size is > 1. Column reporting will get a little confused, though.
* Thread safe.
* On-line querying.
* Historical querying.
* Available as a [reader](https://golang.org/pkg/io/#Reader) middleware, [writer](https://golang.org/pkg/io/#Writer) middleware, or a standalone tracker.

## Example

```go
package examples

import (
	"strings"
	"testing"

	"github.com/erinpentecost/byteline"
)

// TestTerseExample shows off the byteline package.
func TestTerseExample(t *testing.T) {
	// your existing reader
	baseReader := strings.NewReader("Hello!\nThis a string\r\nwith mixed and doubled\n\nnewlines.")
	// is wrapped inside a byteline reader
	tracker := byteline.NewReader(baseReader)
	// and when you read from it
	readBuffer := make([]byte, 60, 60)
	tracker.Read(readBuffer)
	// you can query it at any time:

	// with current offset,
	lastSeenByteOffset, _ := tracker.GetCurrentOffset()
	println(lastSeenByteOffset) // 54
	// with current line and column,
	lastSeenLine, lastSeenColumn, _ := tracker.GetCurrentLineAndColumn()
	println(lastSeenLine)   // 4
	println(lastSeenColumn) // 8
	// with some previously seen offset,
	someLine, someColumn, _ := tracker.GetLineAndColumn(lastSeenByteOffset / 2)
	println(someLine)   // 2
	println(someColumn) // 5
	// or some previously seen line and column.
	someByteOffset, _ := tracker.GetOffset(someLine, someColumn)
	println(someByteOffset) // 27
}
```
