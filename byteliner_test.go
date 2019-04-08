package byteline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOffsetOk(t *testing.T) {
	tracker := newTracker()
	tracker.RunningLineLengths = append(tracker.RunningLineLengths, 10, 20, 30)

	check := func(line, column, expected int) {
		o, e := tracker.GetOffset(line, column)
		assert.NoError(t, e)
		assert.Equal(t, expected, o)
	}

	check(0, 0, 0)
	check(0, 1, 1)
	check(1, 0, 10)
	check(1, 1, 11)
	check(2, 0, 20)
}

func TestGetOffsetError(t *testing.T) {
	tracker := newTracker()
	tracker.RunningLineLengths = append(tracker.RunningLineLengths, 10, 20, 30)

	check := func(line, column int) {
		_, e := tracker.GetOffset(line, column)
		assert.Error(t, e)
	}

	check(-1, 0)
	check(0, -1)

	check(2, 11)
	check(3, 0)

	check(1, 11)
	check(2, 12)
}
