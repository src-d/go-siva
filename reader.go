package iba

import (
	"errors"
	"io"
	"sort"
)

var (
	ErrPendingContent   = errors.New("entry wasn't fully read")
	ErrInvalidCheckshum = errors.New("invalid checksum")
)

type Reader struct {
	r io.ReadSeeker

	current *IndexEntry
	pending uint64
}

func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{r: r}
}

func (r *Reader) Index() (Index, error) {
	endLastBlock, err := r.r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	i, err := r.readIndex(uint64(endLastBlock))
	if err != nil {
		return i, err
	}

	sort.Sort(i)
	return i, nil
}

func (r *Reader) readIndex(offset uint64) (Index, error) {
	i := make(Index, 0)
	if err := i.ReadFrom(r.r, offset); err != nil {
		return nil, err
	}

	if len(i) == 0 || i[0].absStart == 0 {
		return i, nil
	}

	previ, err := r.readIndex(i[0].absStart)
	if err != nil {
		return nil, err
	}

	i = append(i, previ...)
	return i, nil
}

func (r *Reader) Seek(e *IndexEntry) (int64, error) {
	r.current = e
	r.pending = e.Size

	return r.r.Seek(int64(e.absStart), io.SeekStart)
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if r.pending == 0 {
		return 0, io.EOF
	}

	if uint64(len(b)) > r.pending {
		b = b[0:r.pending]
	}

	n, err = r.r.Read(b)
	r.pending -= uint64(n)

	if err == io.EOF && r.pending > 0 {
		err = io.ErrUnexpectedEOF
	}

	return
}
