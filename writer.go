package iba

import (
	"errors"
	"hash"
	"hash/crc32"
	"io"
)

var (
	ErrMissingHeader = errors.New("WriteHeader was not called, or already flushed")
)

type Writer struct {
	w io.Writer

	position uint64
	index    Index
	current  *IndexEntry
	hasher   hash.Hash32
}

func NewWriter(w io.Writer) *Writer {
	crc := crc32.NewIEEE()

	return &Writer{
		w:      io.MultiWriter(w, crc),
		hasher: crc,
	}
}

func (w *Writer) WriteHeader(h *Header) error {
	w.flushIfPending()

	w.current = &IndexEntry{
		Header: (*h),
		Start:  w.position,
	}

	w.index = append(w.index, w.current)
	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	w.position += uint64(n)

	return n, err
}

func (w *Writer) Flush() error {
	if w.current == nil {
		return ErrMissingHeader
	}

	w.current.Size = w.position - w.current.Start
	w.current.CRC32 = w.hasher.Sum32()
	w.hasher.Reset()

	return nil
}

func (w *Writer) flushIfPending() {
	if w.current == nil {
		return
	}

	w.Flush()
}

func (w *Writer) Close() error {
	w.flushIfPending()

	_, err := w.index.WriteTo(w.w)
	return err
}
