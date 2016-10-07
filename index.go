package iba

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sort"
	"time"
)

var (
	IndexSignature = []byte{'I', 'B', 'A'}

	ErrInvalidIndexEntry       = errors.New("invalid index entry")
	ErrInvalidSignature        = errors.New("invalid signature")
	ErrEmptyIndex              = errors.New("empty index")
	ErrUnsupportedIndexVersion = errors.New("unsupported index version")
	ErrCRC32Missmatch          = errors.New("crc32 missmatch")
)

const (
	IndexVersion    uint8 = 1
	indexFooterSize       = 24
)

type Index []*IndexEntry

func (i *Index) ReadFrom(r io.ReadSeeker, endBlock uint64) error {
	if _, err := r.Seek(int64(endBlock)-indexFooterSize, io.SeekStart); err != nil {
		return err
	}

	f, err := i.readFooter(r)
	if err != nil {
		return err
	}

	startingPos := int64(f.IndexSize) + indexFooterSize
	if _, err := r.Seek(-startingPos, io.SeekCurrent); err != nil {
		return err
	}

	defer sort.Sort(i)
	return i.readIndex(r, f, endBlock)
}

func (i *Index) readFooter(r io.Reader) (*IndexFooter, error) {
	f := &IndexFooter{}
	if err := f.ReadFrom(r); err != nil {
		return nil, err
	}

	return f, nil
}

func (i *Index) readIndex(r io.Reader, f *IndexFooter, endBlock uint64) error {
	hr := newHashedReader(r)

	if err := i.readSignature(hr); err != nil {
		return err
	}

	if err := i.readEntries(hr, f, endBlock); err != nil {
		return err
	}

	if f.CRC32 != hr.Checkshum() {
		return ErrCRC32Missmatch
	}

	return nil
}

func (i *Index) readSignature(r io.Reader) error {
	sig := make([]byte, 3)
	if _, err := r.Read(sig); err != nil {
		return err
	}

	if !bytes.Equal(sig, IndexSignature) {
		return ErrInvalidSignature
	}

	var version uint8
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return err
	}

	if version != IndexVersion {
		return ErrUnsupportedIndexVersion
	}

	return nil
}

func (i *Index) readEntries(r io.Reader, f *IndexFooter, endBlock uint64) error {
	for j := 0; j < int(f.EntryCount); j++ {

		e := &IndexEntry{}
		if err := e.ReadFrom(r); err != nil {
			return err
		}

		e.absStart = (endBlock - f.BlockSize) + e.Start
		*i = append(*i, e)
	}

	return nil
}

func (i *Index) WriteTo(w io.Writer) error {
	if len(*i) == 0 {
		return ErrEmptyIndex
	}

	hw := newHashedWriter(w)

	f := &IndexFooter{
		EntryCount: uint32(len(*i)),
	}

	if _, err := hw.Write(IndexSignature); err != nil {
		return err
	}

	if err := binary.Write(hw, binary.BigEndian, IndexVersion); err != nil {
		return err
	}

	var blockSize uint64
	for _, e := range *i {
		blockSize += e.Size
		if err := e.WriteTo(hw); err != nil {
			return err
		}
	}

	f.IndexSize = uint64(hw.Position())
	f.BlockSize = blockSize + f.IndexSize + indexFooterSize
	f.CRC32 = hw.Checkshum()

	if err := f.WriteTo(hw); err != nil {
		return err
	}

	return nil
}

func (s Index) Len() int           { return len(s) }
func (s Index) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Index) Less(i, j int) bool { return s[i].absStart < s[j].absStart }

type IndexEntry struct {
	Header
	Start uint64
	Size  uint64
	CRC32 uint32

	// absStart stores the  abosulute starting position of the entry in the file
	// accross all the blocks in the file, is calculate on-the-fly, so thats
	// why is not stored
	absStart uint64
}

func (e *IndexEntry) WriteTo(w io.Writer) error {
	if e.Name == "" || e.Size == 0 {
		return ErrInvalidIndexEntry
	}

	name := []byte(e.Name)
	length := uint32(len(name))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	if _, err := w.Write(name); err != nil {
		return err
	}

	return writeBinary(w, []interface{}{
		e.Mode,
		e.ModTime.UnixNano(),
		e.Start,
		e.Size,
		e.CRC32,
		e.Flags,
	})
}

func (e *IndexEntry) ReadFrom(r io.Reader) error {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}

	filename := make([]byte, length)
	if _, err := r.Read(filename); err != nil {
		return err
	}

	var nsec int64
	err := readBinary(r, []interface{}{
		&e.Mode,
		&nsec,
		&e.Start,
		&e.Size,
		&e.CRC32,
		&e.Flags,
	})

	e.Name = string(filename)
	e.ModTime = time.Unix(0, nsec)
	return err
}

type IndexFooter struct {
	EntryCount uint32
	IndexSize  uint64
	BlockSize  uint64
	CRC32      uint32
}

func (f *IndexFooter) ReadFrom(r io.Reader) error {
	return readBinary(r, []interface{}{
		&f.EntryCount,
		&f.IndexSize,
		&f.BlockSize,
		&f.CRC32,
	})
}

func (f *IndexFooter) WriteTo(w io.Writer) error {
	return writeBinary(w, []interface{}{
		f.EntryCount,
		f.IndexSize,
		f.BlockSize,
		f.CRC32,
	})
}

func writeBinary(w io.Writer, data []interface{}) error {
	for _, v := range data {
		err := binary.Write(w, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func readBinary(r io.Reader, data []interface{}) error {
	for _, v := range data {
		err := binary.Read(r, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}

	return nil
}
