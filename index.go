package iba

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

var (
	IndexSignature = []byte{'I', 'B', 'A'}

	ErrInvalidIndexEntry = errors.New("invalid index entry")
	ErrInvalidSignature  = errors.New("invalid signature")
)

const indexFooterSize = 20

type Index []*IndexEntry

func (i *Index) Read(r io.ReadSeeker) (n int64, err error) {
	if _, err := r.Seek(-indexFooterSize, io.SeekEnd); err != nil {
		return -1, err
	}

	f, err := i.readFooter(r)
	if err != nil {
		return 0, err
	}

	_, err = r.Seek(-int64(f.Size)-indexFooterSize, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	n += int64(len(IndexSignature))
	if err := i.readSignature(r); err != nil {
		return 0, err
	}

	b, err := i.readEntries(r, f)
	n += b
	if err != nil {
		return 0, err
	}

	return
}

func (i Index) readFooter(r io.Reader) (*IndexFooter, error) {
	f := &IndexFooter{}
	if _, err := f.ReadFrom(r); err != nil {
		return nil, err
	}

	return f, nil
}

func (i Index) readSignature(r io.Reader) error {
	sig := make([]byte, 3)
	if _, err := r.Read(sig); err != nil {
		return err
	}

	if !bytes.Equal(sig, IndexSignature) {
		return ErrInvalidSignature
	}

	return nil
}

func (i Index) readEntries(r io.Reader, f *IndexFooter) (n int64, err error) {
	for j := 0; j < int(f.EntryCount); j++ {
		e := &IndexEntry{}
		b, err := e.ReadFrom(r)
		n += b

		if err != nil {
			return 0, err
		}

		i = append(i, e)
	}

	return
}

func (i Index) WriteTo(w io.Writer) (int64, error) {
	f := &IndexFooter{
		EntryCount: uint32(len(i)),
	}

	var size uint64
	n, err := w.Write(IndexSignature)
	if err != nil {
		return 0, err
	}

	size += uint64(n)

	for _, e := range i {
		n, err := e.WriteTo(w)
		size += uint64(n)

		if err != nil {
			return int64(n), err
		}
	}

	f.Size = size
	return f.WriteTo(w)
}

type Header struct {
	Name    string
	ModTime time.Time
	Mode    os.FileMode
}

type IndexEntry struct {
	Header
	Start uint64
	Size  uint64
	CRC32 uint32
}

var nonVarIndexValues = int64(4) + 8 + 8 + 8 + 4

func (e *IndexEntry) WriteTo(w io.Writer) (n int64, err error) {
	if e.Name == "" || e.Size == 0 {
		return 0, ErrInvalidIndexEntry
	}

	var written int64

	name := []byte(e.Name)
	length := uint32(len(name))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return 0, err
	}

	written += 4 + int64(length) //len of name size and name
	if _, err := w.Write(name); err != nil {
		return 0, err
	}

	var data = []interface{}{
		e.Mode,
		e.ModTime.UnixNano(),
		e.Start,
		e.Size,
		e.CRC32,
	}

	for _, v := range data {
		err := binary.Write(w, binary.BigEndian, v)
		if err != nil {
			return 0, err
		}
	}

	written += nonVarIndexValues //size of data values
	return written, nil
}

func (e *IndexEntry) ReadFrom(r io.Reader) (n int64, err error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return 0, err
	}

	filename := make([]byte, length)
	if _, err := r.Read(filename); err != nil {
		return 0, err
	}

	e.Name = string(filename)

	var nsec int64
	var data = []interface{}{
		&e.Mode,
		&nsec,
		&e.Start,
		&e.Size,
		&e.CRC32,
	}

	for _, v := range data {
		err := binary.Read(r, binary.BigEndian, v)
		if err != nil {
			return 0, err
		}
	}

	e.ModTime = time.Unix(0, nsec)
	return int64(length) + 2 + nonVarIndexValues, nil
}

type IndexFooter struct {
	EntryCount    uint32
	Size          uint64
	CRC32         uint32
	PreviousIndex uint32
}

func (f *IndexFooter) ReadFrom(r io.Reader) (int64, error) {
	var data = []interface{}{
		&f.EntryCount,
		&f.Size,
		&f.CRC32,
		&f.PreviousIndex,
	}

	for _, v := range data {
		err := binary.Read(r, binary.BigEndian, v)
		if err != nil {
			return 0, err
		}
	}

	return indexFooterSize, nil
}

func (f *IndexFooter) WriteTo(w io.Writer) (int64, error) {
	var data = []interface{}{
		f.EntryCount,
		f.Size,
		f.CRC32,
		f.PreviousIndex,
	}

	for _, v := range data {
		err := binary.Write(w, binary.BigEndian, v)
		if err != nil {
			return 0, err
		}
	}

	return indexFooterSize, nil
}
