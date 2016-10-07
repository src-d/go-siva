package iba

import (
	"errors"
	"io"
	"sort"
)

var (
	ErrPendingContent   = errors.New("entry wasn't fully read")
	ErrInvalidCheckshum = errors.New("invalid checksum")
	ErrInvalidReaderAt  = errors.New("reader provided doen't implements ReaderAt interface")
)

// A Reader provides random access to the contents of a shiva archive.
type Reader struct {
	r io.ReadSeeker

	current *IndexEntry
	pending uint64
}

// NewReader creates a new Reader reading from r, reader requires be seekable
// and optionally should implement io.ReaderAt to make usage of the Get method
func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{r: r}
}

// Index reads the index of the shiva file from the provided reader
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

// Get returns a new io.SectionReader allowing concurrent read access to the
// content of the read
func (r *Reader) Get(e *IndexEntry) (*io.SectionReader, error) {
	ra, ok := r.r.(io.ReaderAt)
	if !ok {
		return nil, ErrInvalidReaderAt
	}

	return io.NewSectionReader(ra, int64(e.absStart), int64(e.Size)), nil
}

// Seek seek the internal reader to the starting position of the content for the
// given IndexEntry
func (r *Reader) Seek(e *IndexEntry) (int64, error) {
	r.current = e
	r.pending = e.Size

	return r.r.Seek(int64(e.absStart), io.SeekStart)
}

// Read reads up to len(p) bytes, starting at the current position set by Seek
// and ending in the end of the content, retuning a io.EOF when its reached
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.pending == 0 {
		return 0, io.EOF
	}

	if uint64(len(p)) > r.pending {
		p = p[0:r.pending]
	}

	n, err = r.r.Read(p)
	r.pending -= uint64(n)

	if err == io.EOF && r.pending > 0 {
		err = io.ErrUnexpectedEOF
	}

	return
}
