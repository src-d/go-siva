package siva

import "io"

//ReadWriter can read and write to the same siva file.
//It is not thread-safe.
type ReadWriter struct {
	*reader
	*writer
}

func NewReaderWriter(rw io.ReadWriteSeeker) (*ReadWriter, error) {
	_, ok := rw.(io.ReaderAt)
	if !ok {
		return nil, ErrInvalidReaderAt
	}

	i, err := readIndex(rw)
	if err != nil && err != ErrEmptyIndex {
		return nil, err
	}

	w := newWriter(rw)
	getIndexFunc := func() (Index, error) {
		return append(i, w.index...), nil
	}
	r := newReaderWithIndex(rw, getIndexFunc)
	return &ReadWriter{r, w}, nil
}
