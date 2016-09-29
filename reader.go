package iba

import "io"

type Reader struct {
	r io.ReadSeeker

	current *IndexEntry
	pending uint64
}

func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{r: r}
}

func (r *Reader) Index() (Index, error) {
	i := make(Index, 0)
	_, err := i.Read(r.r)

	return i, err
}

func (r *Reader) Seek(e *IndexEntry) (int64, error) {
	r.current = e
	r.pending = e.Size

	return r.r.Seek(int64(e.Start), io.SeekStart)
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
