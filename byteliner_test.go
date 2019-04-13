package byteline

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkSameOk(t *testing.T, b ByteLiner, line int, col int, offset int) {
	goffset, oerr := b.GetOffset(line, col)
	gline, gcol, lerr := b.GetLineAndColumn(offset)

	if assert.NoError(t, oerr, fmt.Sprintf("GetOffset(%v, %v) returned an error", line, col)) {
		assert.Equal(t, offset, goffset, fmt.Sprintf("GetOffset(%v, %v) is wrong.", line, col))
	}

	if assert.NoError(t, lerr, fmt.Sprintf("GetLineAndCol(%v) returned an error", offset)) {
		assert.Equal(t, line, gline, fmt.Sprintf("GetLineAndCol(%v) returned a bad line.", offset))
		assert.Equal(t, col, gcol, fmt.Sprintf("GetLineAndCol(%v) returned a bad column", offset))
	}

}

func TestSame(t *testing.T) {
	tracker := newTracker()
	tracker.RunningLineLengths = append(tracker.RunningLineLengths, 10, 20, 30)
	// Good values.
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 0, 1, 1)
	checkSameOk(t, tracker, 0, 5, 5)
	checkSameOk(t, tracker, 0, 9, 9)
	checkSameOk(t, tracker, 1, 0, 10)
	checkSameOk(t, tracker, 1, 1, 11)
	checkSameOk(t, tracker, 1, 9, 19)
	checkSameOk(t, tracker, 2, 0, 20)
	checkSameOk(t, tracker, 2, 1, 21)
	checkSameOk(t, tracker, 2, 9, 29)

	checkSameOk(t, tracker, 3, 0, 30)
}

func TestGetOffsetError(t *testing.T) {
	tracker := newTracker()
	tracker.RunningLineLengths = append(tracker.RunningLineLengths, 10, 20, 30)

	check := func(line, column int) {
		_, e := tracker.GetOffset(line, column)
		assert.NotNil(t, e)
	}

	check(-1, 0)
	check(0, -1)

	check(2, 11)

	check(1, 11)
	check(2, 12)

}

func TestGetLineColError(t *testing.T) {
	tracker := newTracker()
	tracker.RunningLineLengths = append(tracker.RunningLineLengths, 10, 20, 30)

	check := func(offset int) {
		_, _, e := tracker.GetLineAndColumn(offset)
		assert.NotNil(t, e)
	}

	check(-1)
	check(31)
}
