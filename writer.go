package iba

import (
	"errors"
	"io"
)

var (
	ErrMissingHeader = errors.New("WriteHeader was not called, or already flushed")
	ErrClosedWriter  = errors.New("Writer is closed")
)

type Writer struct {
	w        *hashedWriter
	index    Index
	current  *IndexEntry
	position uint64
	closed   bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: newHashedWriter(w),
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
	w.current.CRC32 = w.w.Checkshum()
	w.w.Reset()

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

	err := w.index.WriteTo(w.w)
	if err == ErrEmptyIndex {
		return nil
	}

	return err
}
