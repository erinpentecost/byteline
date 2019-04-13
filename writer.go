package byteline

import (
	"io"
)

type Writer struct {
	err   error
	track *tracker
	wr    io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		err:   nil,
		track: newTracker(),
		wr:    w,
	}
}

// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
func (w *Writer) Write(p []byte) (n int, err error) {
	return
}

func (w *Writer) WriteString(s string) (n int, err error) {
	return
}
