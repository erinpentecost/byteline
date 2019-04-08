package byteline

import (
	"fmt"
	"sort"
)

// ByteLiner supports mapping between byte offset and line/column.
// Lines start at 0, columns start at 0.
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
		RunningLineLengths: make([]int, 1, 500),
	}
	t.RunningLineLengths[0] = 0
	return t
}

func (t *tracker) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	// Step 1: Search RunningLineLengths for where byteOffset would fall into.
	// Step 2: Save that index: it's the line.
	// Step 3: byteOffset - RunningLineLengths[line] is positive, and is the column.

	line = sort.SearchInts(t.RunningLineLengths, byteOffset)
	if line < len(t.RunningLineLengths) {
		col = byteOffset - t.RunningLineLengths[line+1]
	} else {
		ok = fmt.Errorf("requested byteOffset %v is beyond the last seen offset %v",
			byteOffset,
			t.RunningLineLengths[len(t.RunningLineLengths)-1])
	}

	// internally, lines start at 1
	line--

	return
}

func (t *tracker) GetOffset(line int, column int) (offset int, ok error) {
	// Step 1: Return RunningLineLengths[line] + column.
	// If that's over RunningLineLengths[line + 1], the caller made a boo-boo.

	if line < 0 {
		ok = fmt.Errorf("by convention, the first line is 0. %v is before that", line)
		return
	}
	// internally, lines start at 1.
	line++

	if column < 0 {
		ok = fmt.Errorf("by convention, the first column is 0. %v is before that", column)
		return
	}
	if line >= len(t.RunningLineLengths) {
		ok = fmt.Errorf("requested line %v is beyond the last seen line %v", line, len(t.RunningLineLengths))
		return
	}

	offset = t.RunningLineLengths[line-1] + column
	if offset > t.RunningLineLengths[line] {
		ok = fmt.Errorf("requested column %v is beyond the end (%v) of requested line %v",
			column,
			t.RunningLineLengths[line],
			line)
	}

	return
}
