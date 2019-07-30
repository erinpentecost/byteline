package byteline

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestLFFileReader(t *testing.T) {
	testRead := func(f string, testRune rune) {
		// get a handle on the file
		absPath, err := filepath.Abs(f)
		testFile, err := ioutil.ReadFile(absPath)
		testFileString := string(testFile)
		assert.NoError(t, err, f)
		// get real number of newlines
		newlines := 0
		for _, r := range testFileString {
			if r == testRune {
				newlines++
			}
		}
		// run the tracker
		trackReader := NewReader(strings.NewReader(testFileString))
		// read all bytes
		readBuffer := make([]byte, 60, 60)
		err = nil
		for {
			count, err := trackReader.Read(readBuffer)
			if count < 1 || err != nil {
				break
			}
		}
		// confirm we got the right number of newlines
		fline, _, ferr := trackReader.GetCurrentLineAndColumn()
		assert.NoError(t, ferr)
		assert.Equal(t, newlines, fline, f)
	}

	err := filepath.Walk("./testdata/", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasPrefix(info.Name(), "lf") {
				testRead(path, '\n')
			} else if strings.HasPrefix(info.Name(), "cr") {
				testRead(path, '\r')
			}
		}
		return nil
	})
	assert.NoError(t, err)
}
