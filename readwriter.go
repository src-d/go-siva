package siva

import "io"

//ReadWriter can read and write to the same siva file.
//It is not thread-safe.
type ReadWriter struct {
	*Reader
	*Writer
}

func NewReaderWriter(rw io.ReadWriteSeeker) (*ReadWriter, error) {
	_, ok := rw.(io.ReaderAt)
	if !ok {
		return nil, ErrInvalidReaderAt
	}
	w := NewWriter(rw)
	getIndexFunc := func() (Index, error) {
		return w.index, nil
	}
	r := newReaderWithIndex(rw, getIndexFunc)
	return &ReadWriter{r, w}, nil
}
