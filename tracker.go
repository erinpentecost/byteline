package byteline

import (
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"unicode/utf8"
)

// Tracker keeps track of line end offsets.
type Tracker struct {
	// Only completed lines are allowed in this data structure.
	// A line's last newline character's index is the value
	// for that line. Lines start at 0.
	lineEndIndices           []int
	currentLineLastSeenIndex int
	buf                      []byte
	err                      error
	prev                     rune
	mux                      sync.Mutex
}

// NewTracker creates a new Tracker.
func NewTracker() *Tracker {
	t := &Tracker{
		lineEndIndices:           make([]int, 0, 500),
		currentLineLastSeenIndex: -1,
		buf:                      make([]byte, 0, 4),
		prev:                     0,
	}
	return t
}

// MarkBytes updates the tracker. Partial runes are ok.
func (t *Tracker) MarkBytes(p []byte) (int, error) {
	// This just exists to buffer up bytes until we
	// see a complete rune. When that happens, the
	// rune (and its size) are marked up.
	t.mux.Lock()
	defer t.mux.Unlock()
	// if it's hosed, give up.
	if t.err != nil {
		return 0, t.err
	}

	// stick buff onto first part of incoming bytes
	var incoming []byte
	if len(t.buf) == 0 {
		incoming = p
	} else {
		incoming := make([]byte, len(p)+len(t.buf))
		copy(t.buf, incoming)
		copy(p, incoming[len(t.buf):])
	}

	// clear buff, we captured them in incoming
	t.buf = make([]byte, 0, 4)

	// start iterating on everything
	i := 0
	for {
		// quit if we get to the end
		if i >= len(incoming) {
			// report the correct number of bytes we read in
			return i, nil
		}
		// get the rune and size of the rune from the input
		r, s := utf8.DecodeRune(incoming[i:])
		if r == utf8.RuneError {
			if s == 0 {
				// we reached the end, which is good.
				break
			}
			// the rune can't be decoded correctly.
			// maybe we only got the first half of the rune?
			// save what's left in the buffer so we can try again later.
			// yes, this will cause the buffer to explode and the tracker to stop
			// tracking if we get a bad byte that never resolves.
			t.buf = incoming[i:]
			oops := fmt.Errorf("can't decode bytes %s into unicode rune", printHead(incoming[i:]))
			// if the error is definitely not recoverable, save to t.err
			if len(t.buf) > 4 {
				t.err = oops
			}
			return i, oops
		}
		// at this point, we have a valid rune and its size
		t.markRune(r, s)
		i += s
	}

	// success
	return i, t.err
}

func (t *Tracker) addToCurrentLine(size int) {
	t.currentLineLastSeenIndex += size
}

func (t *Tracker) endLine() {
	t.lineEndIndices = append(t.lineEndIndices, t.currentLineLastSeenIndex)
}

func (t *Tracker) markRune(r rune, size int) {
	last := t.prev
	t.prev = r

	if last == '\r' || last == '\n' {
		if r == '\r' || r == '\n' {
			// There were two line end runes in a row.
			// Don't continue to collapse them, so mark
			// prev as garbage.
			t.prev = 0

			if last != r {
				// current is end of this line. this the second character
				// of a /r/n or /n/r pair.
				t.addToCurrentLine(size)
				t.endLine()
			} else {
				// previous was an end, and current is also. this is the
				// second character of a /r/r or /n/n pair.
				t.endLine()
				t.addToCurrentLine(size)
				t.endLine()
			}
		} else {
			// previous was an end, but this is not. this is the second
			// character of a /rX or /nX pair.
			t.endLine()
			t.addToCurrentLine(size)
		}
	} else {
		// normal.
		t.addToCurrentLine(size)
	}
}

func printHead(p []byte) string {
	length := 4
	if len(p) < length {
		length = len(p)
	}
	if length == 0 {
		return "<empty>"
	}
	return fmt.Sprintf("<%s>", hex.EncodeToString(p[0:length]))
}

// GetLineAndColumn returns the line and column given a byte offset.
func (t *Tracker) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if byteOffset < 0 {
		ok = fmt.Errorf("valid byteOffset is >= 0, not %v", byteOffset)
		return
	}

	if byteOffset > t.currentLineLastSeenIndex {
		ok = fmt.Errorf("requested byteOffset %v is beyond the last seen byte %v",
			byteOffset,
			t.currentLineLastSeenIndex)
		return
	}

	line = sort.SearchInts(t.lineEndIndices, byteOffset)
	if line == 0 {
		col = byteOffset
	} else {
		col = byteOffset - t.lineEndIndices[line-1] - 1
	}
	return
}

// GetOffset returns the byte offset given a line and column.
func (t *Tracker) GetOffset(line int, column int) (offset int, ok error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	// input validation

	if line < 0 {
		ok = fmt.Errorf("by convention, the first line is 0. %v is before that", line)
		return
	}

	if line > len(t.lineEndIndices) {
		ok = fmt.Errorf("requested line %v is beyond the last seen line %v", line, len(t.lineEndIndices))
		return
	}

	if column < 0 {
		ok = fmt.Errorf("by convention, the first column is 0. %v is before that", column)
		return
	}

	// get sane bounds on the indices of the requested line

	lineStart := 0
	if line > 0 {
		lineStart = t.lineEndIndices[line-1]
		column++ // first line is weird
	}

	lineEnd := t.currentLineLastSeenIndex
	if line < len(t.lineEndIndices) {
		lineEnd = t.lineEndIndices[line]
	}

	// calculate the offset

	offset = column + lineStart

	// check if the offset violated our upper bound

	if offset > lineEnd {
		ok = fmt.Errorf("requested column %v is beyond the end of requested line %v",
			column,
			line)
	}

	return
}

// GetCurrentLineAndColumn returns the current line and column.
func (t *Tracker) GetCurrentLineAndColumn() (line int, col int, ok error) {
	t.mux.Lock()
	defer t.mux.Unlock()
	line, col, ok = t.GetLineAndColumn(t.currentLineLastSeenIndex)
	return
}

// GetCurrentOffset returns the current byte offset.
func (t *Tracker) GetCurrentOffset() (offset int, ok error) {
	t.mux.Lock()
	defer t.mux.Unlock()
	offset = t.currentLineLastSeenIndex
	return
}
