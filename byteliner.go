package byteline

import (
	"fmt"
	"sort"
)

// ByteLiner supports mapping between byte offset and line/column.
// Lines start at 0, columns start at 0.
// The newline character that ends a line is the last column
// on the line it ends.
type ByteLiner interface {
	GetLineAndColumn(byteOffset int) (line int, col int, ok error)
	GetOffset(line int, column int) (offset int, ok error)
}

type tracker struct {
	// RunningLineLengths is an additive line length tracker.
	// For example, a document with 3 lines all of length 10
	// would result in {0,10,20,30}.
	RunningLineLengths []int
}

func newTracker() *tracker {
	t := &tracker{
		RunningLineLengths: make([]int, 0, 500),
	}
	//t.RunningLineLengths[0] = 0
	return t
}

func (t *tracker) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {

	if byteOffset < 0 {
		ok = fmt.Errorf("valid byteOffset is >= 0, not %v", byteOffset)
		return
	}

	line = sort.SearchInts(t.RunningLineLengths, byteOffset)

	if line >= len(t.RunningLineLengths) {
		ok = fmt.Errorf("requested byteOffset %v is beyond the last seen line %v",
			byteOffset,
			len(t.RunningLineLengths)-1)
		return
	}

	lineEnd := t.RunningLineLengths[line]

	if lineEnd == byteOffset {
		line++
		col = 0
		return
	}

	lineStart := 0
	if line > 0 {
		lineStart = t.RunningLineLengths[line-1]
	}

	col = byteOffset - lineStart

	return
}

func (t *tracker) GetOffset(line int, column int) (offset int, ok error) {

	if line < 0 {
		ok = fmt.Errorf("by convention, the first line is 0. %v is before that", line)
		return
	}

	if column < 0 {
		ok = fmt.Errorf("by convention, the first column is 0. %v is before that", column)
		return
	}

	lineStart := 0
	if line > 0 {
		lineStart = t.RunningLineLengths[line-1]
	}

	offset = lineStart + column

	if line >= len(t.RunningLineLengths) && !(line == len(t.RunningLineLengths) && column == 0) {
		ok = fmt.Errorf("requested column %v is beyond the end of requested line %v",
			column,
			line)
	}

	return
}
