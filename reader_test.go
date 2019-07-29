package byteline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnixNewlineReader(t *testing.T) {
	// set up middleware reader
	baseReader := strings.NewReader("Hello There\rPerson")
	trackReader := NewReader(baseReader)
	// read all bytes
	readBuffer := make([]byte, 60, 60)
	_, err := trackReader.Read(readBuffer)
	// make sure we got the correct results.
	assert.Nil(t, err)
	checkSameOk(t, trackReader, 0, 0, 0)
	checkSameOk(t, trackReader, 1, 0, 12)
	checkSameOk(t, trackReader, 1, 4, 16)
}
