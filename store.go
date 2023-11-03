package main

import (
	"io"
	"os"
	"strconv"
	"sync"
)

type Storer interface {
	Read() (uint64, error)
	Write(value uint64) error
}

type MemStore struct {
	mu    sync.RWMutex
	value uint64
}

func NewMemStore() *MemStore {
	return &MemStore{}
}

func (m *MemStore) Read() (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value, nil
}

func (m *MemStore) Write(value uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = value

	return nil
}

type FileStore struct {
	mu    sync.RWMutex
	fname string
}

func NewFileStore(fname string) *FileStore {
	return &FileStore{
		fname: fname,
	}
}

func (f *FileStore) Read() (uint64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, err := os.Stat(f.fname)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0, err
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

	num, err := strconv.ParseUint(string(cnt), 10, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (f *FileStore) Write(value uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

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
