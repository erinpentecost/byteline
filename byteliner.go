package byteline

import (
	"encoding/hex"
	"fmt"
	"hex"
	"sort"
	"unicode/utf8"
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
	buf                []byte
	err                error
	// prevCR is \r
	prevCR bool
	// prevLF is \n
	prevLF            bool
	currentLineLength int
}

func newTracker() *tracker {
	t := &tracker{
		RunningLineLengths: make([]int, 0, 500),
		buf:                make([]byte, 0, 4),
		prevCR:             false,
		prevLF:             false,
		currentLineLength:  0,
	}
	//t.RunningLineLengths[0] = 0
	return t
}

func (t *tracker) markBytes(p []byte) (int, error) {
	// stick buff onto first part of incoming bytes
	var incoming []byte
	if len(t.buf) == 0 {
		incoming = p
	} else {
		incoming := make([]byte, len(p)+len(t.buf), len(p)+len(t.buf))
		copy(t.buf, incoming)
		copy(p, incoming[len(t.buf)])
	}

	// clear buff, we captured them in incoming
	t.buf = make([]byte, 0, 4)

	// start iterating on everything
	i := 0
	for {
		// quit if we get to the end
		if i >= len(incoming) {
			// report the correct number of bytes we read in
			return i - 1, nil
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
		markRune(r, s)
	}
}

// TODO: handle mixed newlines
func (t *tracker) markRune(r rune, size int) {

	if r != '\r' && r != '\n' {
		if t.prevCR || t.prevLF {
			// last char was the end of the line, current one is a new line.
		} else {
			// increment current line
		}
		t.prevCR = false
		t.prevLF = false
	} else {
		// we are close to ending the line.
		if r == '\r' && t.prevCR {
			// true end of the line
			t.prevCR = false
			t.prevLF = false
		} else if r == '\n' && t.prevLF {
			// true end of the line
			t.prevCR = false
			t.prevLF = false
		}
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

	if len(t.RunningLineLengths) > line && offset > t.RunningLineLengths[line] {
		ok = fmt.Errorf("requested column %v is beyond the end of requested line %v",
			column,
			line)
	}

	return
}
