package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Storer interface {
	Read() (uint64, error)
	Write(value uint64) error
}

type NoopStore struct{}

func NewNoopStore() *NoopStore {
	return &NoopStore{}
}

func (m *NoopStore) Read() (uint64, error) {
	return uint64(0), nil
}

func (m *NoopStore) Write(value uint64) error {
	return nil
}

type FileStore struct {
	fname string
}

func NewFileStore(fname string) *FileStore {
	return &FileStore{
		fname: fname,
	}
}

func (f *FileStore) Read() (uint64, error) {
	_, err := os.Stat(f.fname)
	if err != nil {
		return 0, fmt.Errorf("FileStore Read: %w", err)
	}

	file, err := os.Open(f.fname)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return f.readCnt(file)
}

func (f *FileStore) readCnt(r io.Reader) (uint64, error) {
	cnt, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseUint(strings.TrimRight(string(cnt), "\n"), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (f *FileStore) Write(value uint64) error {
	file, err := os.Create(f.fname)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.writeCnt(file, value)
}

func (f *FileStore) writeCnt(w io.Writer, v uint64) error {
	_, err := w.Write([]byte(strconv.FormatUint(v, 10)))
	if err != nil {
		return err
	}
	return nil
}
