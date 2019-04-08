package byteline

type ByteLiner interface {
	GetLineAndColumn(byteOffset int) (line int, col int, ok error)
	GetOffset(line int, column int) (offset int, ok error)
}

type Tracker struct {
	// RunningLineLengths is an additive line length tracker.
	// For example, a document with 3 lines all of length 10
	// would result in {10,20,30}.
	RunningLineLengths []int
}

func (t *Tracker) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	// Step 1: Search RunningLineLengths for where byteOffset would fall into.
	// Step 2: Save that index to i. This is the line.
	// Step 3: byteOffset - RunningLineLengths[i] is positive, and is the column.

}

func (t *Tracker) GetOffset(line int, column int) (offset int, ok error) {
	// Step 1: Return RunningLineLengths[line] + column.
	// If that's over RunningLineLengths[line + 1], the caller made a boo-boo.
}
