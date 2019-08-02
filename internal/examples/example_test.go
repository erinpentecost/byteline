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
