package byteline

import (
	"io"
)

// Reader wraps around an internal reader and counts lines.
type Reader struct {
	track *Tracker
	re    io.Reader
}

// NewReader returns a new byteline reader middleware.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		track: NewTracker(),
		re:    r,
	}
}

// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered. Even if Read
// returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read. It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil. The next Read should
// return 0, EOF.
//
// Callers should always process the n > 0 bytes returned before
// considering the error err. Doing so correctly handles I/O errors
// that happen after reading some bytes and also both of the
// allowed EOF behaviors.
//
// Implementations of Read are discouraged from returning a
// zero byte count with a nil error, except when len(p) == 0.
// Callers should treat a return of 0 and nil as indicating that
// nothing happened; in particular it does not indicate EOF.
func (r *Reader) Read(p []byte) (n int, err error) {

	n, err = r.re.Read(p)
	_, trackErr := r.track.MarkBytes(p[:n])

	if trackErr != nil && err == nil {
		err = trackErr
	}

	return
}

// TrackError is non-nil if the byteline tracker encountered an
// unrecoverable error.
func (r *Reader) TrackError() error {
	return r.track.err
}

// GetLineAndColumn returns the line and column for the given byte offset.
func (r *Reader) GetLineAndColumn(byteOffset int) (line int, col int, ok error) {
	return r.track.GetLineAndColumn(byteOffset)
}

// GetOffset returns the byte offset for the given line and column.
func (r *Reader) GetOffset(line int, column int) (offset int, ok error) {
	return r.track.GetOffset(line, column)
}

// GetCurrentLineAndColumn returns the current line and column.
func (r *Reader) GetCurrentLineAndColumn() (line int, col int, ok error) {
	return r.track.GetCurrentLineAndColumn()
}

// GetCurrentOffset returns the current byte offset.
func (r *Reader) GetCurrentOffset() (offset int, ok error) {
	return r.track.GetCurrentOffset()
}
