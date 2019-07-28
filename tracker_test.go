package byteline

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkSameOk(t *testing.T, b ByteLiner, line int, col int, offset int) {
	goffset, oerr := b.GetOffset(line, col)
	gline, gcol, lerr := b.GetLineAndColumn(offset)

	if assert.NoError(t, oerr, fmt.Sprintf("GetOffset(%v, %v) returned an error", line, col)) {
		assert.Equal(t, offset, goffset, fmt.Sprintf("GetOffset(%v, %v) is wrong", line, col))
	}

	if assert.NoError(t, lerr, fmt.Sprintf("GetLineAndCol(%v) returned an error", offset)) {
		assert.Equal(t, line, gline, fmt.Sprintf("GetLineAndCol(%v) returned a bad line", offset))
		assert.Equal(t, col, gcol, fmt.Sprintf("GetLineAndCol(%v) returned a bad column", offset))
	}

}

func TestSearch(t *testing.T) {
	testList := []int{0, 3, 5, 7}
	assert.Equal(t, 0, sort.SearchInts(testList, 0))
	assert.Equal(t, 1, sort.SearchInts(testList, 1))
	assert.Equal(t, 1, sort.SearchInts(testList, 3))
}

func TestSame(t *testing.T) {
	tracker := NewTracker()
	tracker.lineEndIndices = []int{10, 20, 30}
	tracker.currentLineLastSeenIndex = 100
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 0, 1, 1)
	checkSameOk(t, tracker, 0, 5, 5)
	checkSameOk(t, tracker, 0, 9, 9)
	checkSameOk(t, tracker, 0, 10, 10)

	checkSameOk(t, tracker, 1, 0, 11)
	checkSameOk(t, tracker, 1, 9, 20)

	checkSameOk(t, tracker, 2, 0, 21)
	checkSameOk(t, tracker, 2, 1, 22)
	checkSameOk(t, tracker, 2, 9, 30)
}

func TestEmpty(t *testing.T) {
	tracker := NewTracker()
	tracker.lineEndIndices = []int{0, 1, 2, 3, 4}
	tracker.currentLineLastSeenIndex = 100
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 1, 0, 1)
	checkSameOk(t, tracker, 2, 0, 2)
	checkSameOk(t, tracker, 3, 0, 3)
}

func TestOneRune(t *testing.T) {
	tracker := NewTracker()
	text := "e\nr\ni\nn"
	n, err := tracker.MarkBytes([]byte(text))
	assert.Nil(t, err)
	assert.Equal(t, len(text), n)
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 0, 1, 1)
	checkSameOk(t, tracker, 1, 0, 2)
	checkSameOk(t, tracker, 1, 1, 3)
	checkSameOk(t, tracker, 2, 0, 4)
	checkSameOk(t, tracker, 2, 1, 5)
	// can we handle current line?
	checkSameOk(t, tracker, 3, 0, 6)
}

func TestGetOffsetError(t *testing.T) {
	tracker := NewTracker()
	tracker.lineEndIndices = append(tracker.lineEndIndices, 10, 20, 30)
	tracker.currentLineLastSeenIndex = 100

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
	tracker := NewTracker()
	tracker.lineEndIndices = append(tracker.lineEndIndices, 10, 20, 30)
	tracker.currentLineLastSeenIndex = 30

	check := func(offset int) {
		_, _, e := tracker.GetLineAndColumn(offset)
		assert.NotNil(t, e)
	}

	check(-1)
	check(31)
}

func TestUnixNewline(t *testing.T) {
	tracker := NewTracker()
	text := "Hello There\rPerson"
	n, err := tracker.MarkBytes([]byte(text))
	assert.Nil(t, err)
	assert.Equal(t, len(text), n)
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 1, 0, 12)
	checkSameOk(t, tracker, 1, 4, 16)
}

func TestWindowsNewline(t *testing.T) {
	tracker := NewTracker()
	text := "Hello There\r\nPerson"
	n, err := tracker.MarkBytes([]byte(text))
	assert.Nil(t, err)
	assert.Equal(t, len(text), n)
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 1, 0, 13)
	checkSameOk(t, tracker, 1, 4, 17)
}

func TestDoubleUnixNewline(t *testing.T) {
	tracker := NewTracker()
	text := "Hello There\r\rPerson"
	n, err := tracker.MarkBytes([]byte(text))
	assert.Nil(t, err)
	assert.Equal(t, len(text), n)
	checkSameOk(t, tracker, 0, 0, 0)
	checkSameOk(t, tracker, 0, 11, 11)
	checkSameOk(t, tracker, 1, 0, 12)
	checkSameOk(t, tracker, 2, 0, 13)
	checkSameOk(t, tracker, 2, 1, 14)
}
