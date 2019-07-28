package byteline

import (
	"io"
)

// Writer wraps around an internal writer and counts lines.
type Writer struct {
	track *Tracker
	wr    io.Writer
}

// NewWriter returns a new byteline writer middleware.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		track: NewTracker(),
		wr:    w,
	}
}

// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
func (w *Writer) Write(p []byte) (n int, err error) {

	n, err = w.wr.Write(p)
	_, trackErr := w.track.MarkBytes(p[:n])

	if trackErr != nil && err == nil {
		err = trackErr
	}

	return
}

// TrackError is non-nil if the byteline tracker encountered an
// unrecoverable error.
func (w *Writer) TrackError() error {
	return w.track.err
}

// GetLineAndColumn returns the line and column for the given byte offset.
func (w *Writer) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	return w.track.GetLineAndColumn(byteOffset)
}

// GetOffset returns the byte offset for the given line and column.
func (w *Writer) GetOffset(line int, column int) (offset int, ok error) {
	return w.track.GetOffset(line, column)
}
