package byteline

import (
	"io"
)

type Writer struct {
	track *tracker
	wr    io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		track: newTracker(),
		wr:    w,
	}
}

// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
func (w *Writer) Write(p []byte) (n int, err error) {

	n, err = w.wr.Write(p)
	_, trackErr := w.track.markBytes(p[:n])

	if trackErr != nil && err == nil {
		err = trackErr
	}

	return
}

func (w *Writer) TrackError() error {
	return w.track.err
}

func (w *Writer) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	return w.track.GetLineAndColumn(byteOffset)
}
func (w *Writer) GetOffset(line int, column int) (offset int, ok error) {
	return w.track.GetOffset(line, column)
}
