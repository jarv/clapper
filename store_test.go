package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReadNoExist(t *testing.T) {
	f := NewFileStore("/file/does/not/exist")
	_, err := f.Read()
	assert.Error(t, err, "Expected an error to occur")
}

func TestFileRead(t *testing.T) {
	f := NewFileStore("/some/file")
	v, err := f.readCnt(strings.NewReader("9999"))
	require.NoError(t, err)
	assert.Equal(t, uint64(9999), v)
}

func TestFileWrite(t *testing.T) {
	f := NewFileStore("/some/file")
	var b bytes.Buffer
	require.NoError(t, f.writeCnt(&b, uint64(9999)))
	assert.Equal(t, "9999", b.String())
}
