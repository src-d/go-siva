package iba

import (
	"bytes"
	"sort"
	"time"

	. "gopkg.in/check.v1"
)

type IndexSuite struct{}

var _ = Suite(&IndexSuite{})

func (s *IndexSuite) TestIndexWriteToEmpty(c *C) {
	i := make(Index, 0)
	err := i.WriteTo(nil)
	c.Assert(err, Equals, ErrEmptyIndex)
}

func (s *IndexSuite) TestIndexSort(c *C) {
	i := Index{{absStart: 100}, {absStart: 10}}
	sort.Sort(i)

	c.Assert(int(i[0].absStart), Equals, 10)
	c.Assert(int(i[1].absStart), Equals, 100)
}

func (s *IndexSuite) TestIndexFooterIdempotent(c *C) {
	expected := &IndexFooter{
		EntryCount: 2,
		IndexSize:  42,
		BlockSize:  84,
		CRC32:      4242,
	}

	buf := bytes.NewBuffer(nil)
	err := expected.WriteTo(buf)
	c.Assert(err, IsNil)

	footer := &IndexFooter{}
	err = footer.ReadFrom(buf)
	c.Assert(err, IsNil)
	c.Assert(footer, DeepEquals, expected)
}

func (s *IndexSuite) TestIndexEntryIdempotent(c *C) {
	expected := &IndexEntry{}
	expected.Name = "foo"
	expected.Mode = 0644
	expected.ModTime = time.Now()
	expected.Start = 84
	expected.Size = 42
	expected.CRC32 = 4242
	expected.Flags = FlagDeleted

	buf := bytes.NewBuffer(nil)
	err := expected.WriteTo(buf)
	c.Assert(err, IsNil)

	entry := &IndexEntry{}
	err = entry.ReadFrom(buf)
	c.Assert(err, IsNil)
	c.Assert(entry, DeepEquals, expected)
}

func (s *IndexSuite) TestIndexIdempotent(c *C) {
	e := &IndexEntry{}
	e.Name = "foo"
	e.Mode = 0644
	e.ModTime = time.Now()
	e.Start = 10
	e.Size = 3
	e.CRC32 = 42
	e.absStart = 10

	expected := make(Index, 0)
	expected = append(expected, e)

	buf := bytes.NewBuffer([]byte("foo"))
	err := expected.WriteTo(buf)
	c.Assert(err, IsNil)

	index := make(Index, 0)
	r := bytes.NewReader(buf.Bytes())
	err = index.ReadFrom(r, uint64(buf.Len()))
	c.Assert(err, IsNil)
	c.Assert(index, DeepEquals, expected)
	c.Assert(index, HasLen, 1)
}
