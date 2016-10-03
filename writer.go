package iba

import (
	"errors"
	"hash"
	"hash/crc32"
	"io"
)

var (
	ErrMissingHeader = errors.New("WriteHeader was not called, or already flushed")
	ErrClosedWriter  = errors.New("Writer is closed")
)

type Writer struct {
	w        io.Writer
	position uint64
	previous uint64
	index    Index
	current  *IndexEntry
	hasher   hash.Hash32
	closed   bool
}

func NewWriter(w io.Writer, size uint64) *Writer {
	crc := crc32.NewIEEE()

	return &Writer{
		w:        io.MultiWriter(w, crc),
		position: size,
		previous: size,
		hasher:   crc,
	}
}

func (w *Writer) WriteHeader(h *Header) error {
	if err := w.flushIfPending(); err != nil {
		return err
	}

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
	if w.closed {
		return ErrClosedWriter
	}

	if w.current == nil {
		return ErrMissingHeader
	}

	w.current.Size = w.position - w.current.Start
	w.current.CRC32 = w.hasher.Sum32()
	w.hasher.Reset()

	return nil
}

func (w *Writer) flushIfPending() error {
	if w.closed {
		return ErrClosedWriter
	}

	if w.current == nil {
		return nil
	}

	return w.Flush()
}

func (w *Writer) Close() error {
	defer func() { w.closed = true }()

	if err := w.flushIfPending(); err != nil {
		return err
	}

	return w.index.WriteTo(w.w, w.previous)
}
