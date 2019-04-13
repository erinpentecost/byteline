package byteline

import (
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"unicode/utf8"
)

type Tracker struct {
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
	mux               sync.Mutex
}

func NewTracker() *Tracker {
	t := &Tracker{
		RunningLineLengths: make([]int, 0, 500),
		buf:                make([]byte, 0, 4),
		prevCR:             false,
		prevLF:             false,
		currentLineLength:  0,
	}
	//t.RunningLineLengths[0] = 0
	return t
}

func (t *Tracker) MarkBytes(p []byte) (int, error) {
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

func (t *Tracker) markRune(r rune, size int) {

	if r != '\r' && r != '\n' {
		if t.prevCR || t.prevLF {
			// last char was the end of the line, current one is a new line.
			t.RunningLineLengths = append(t.RunningLineLengths, t.currentLineLength)
			t.currentLineLength = size
		} else {
			// increment current line
			t.currentLineLength += size
		}
		t.prevCR = false
		t.prevLF = false
	} else {
		t.currentLineLength += size

		// we are close to ending the line.
		if r == '\r' && t.prevCR {
			// true end of the line in a two-rune line ending (Windows)
			t.RunningLineLengths = append(t.RunningLineLengths, t.currentLineLength)
			t.currentLineLength = 0
			t.prevCR = false
			t.prevLF = false
		} else if r == '\n' && t.prevLF {
			// true end of the line in a two-rune line ending (Acorn BBC + RISC OS)
			t.RunningLineLengths = append(t.RunningLineLengths, t.currentLineLength)
			t.currentLineLength = 0
			t.prevCR = false
			t.prevLF = false
		} else if r == '\r' && t.prevLF {
			// two \r's in a row
			t.RunningLineLengths = append(t.RunningLineLengths, t.currentLineLength, 1)
			t.currentLineLength = 0
			t.prevCR = false
			t.prevLF = false
		} else if r == '\n' && t.prevCR {
			// two \n's in a row
			t.RunningLineLengths = append(t.RunningLineLengths, t.currentLineLength, 1)
			t.currentLineLength = 0
			t.prevCR = false
			t.prevLF = false
		} else if r == '\r' {
			// this is the start of an end-of-line chord
			t.prevCR = true
			t.prevLF = false
		} else if r == '\n' {
			// this is the start of an end-of-line chord
			t.prevCR = false
			t.prevLF = true
		} else {
			panic("oh no")
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

func (t *Tracker) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if byteOffset < 0 {
		ok = fmt.Errorf("valid byteOffset is >= 0, not %v", byteOffset)
		return
	}

	line = sort.SearchInts(t.RunningLineLengths, byteOffset)

	if line == len(t.RunningLineLengths) &&
		(byteOffset <= t.RunningLineLengths[line-1]+t.currentLineLength) {
		col = byteOffset - t.RunningLineLengths[line-1]
		return
	} else if line >= len(t.RunningLineLengths) {
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

func (t *Tracker) GetOffset(line int, column int) (offset int, ok error) {
	t.mux.Lock()
	defer t.mux.Unlock()

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