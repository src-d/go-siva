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
	if err == ErrEmptyIndex {
		i = Index{}
	} else if err != nil {
		return nil, err
	}

	w := newWriter(rw)
	getIndexFunc := func() (Index, error) {
		index := Index{}
		index = append(index, i...)
		index = append(index, w.index...)
		return index, nil
	}
	r := newReaderWithIndex(rw, getIndexFunc)
	return &ReadWriter{r, w}, nil
}
