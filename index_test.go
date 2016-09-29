package iba

import (
	"bytes"
	"time"

	. "gopkg.in/check.v1"
)

type IndexSuite struct{}

var _ = Suite(&IndexSuite{})

func (s *IndexSuite) TestIndexFooterIdempotent(c *C) {
	expected := &IndexFooter{
		EntryCount:    2,
		Size:          42,
		CRC32:         4242,
		PreviousIndex: 84,
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
	e.Start = 84
	e.Size = 42
	e.CRC32 = 4242

	expected := make(Index, 0)
	expected = append(expected, e)

	buf := bytes.NewBuffer(nil)
	err := expected.WriteTo(buf)
	c.Assert(err, IsNil)

	index := make(Index, 0)
	r := bytes.NewReader(buf.Bytes())
	err = index.ReadFrom(r)
	c.Assert(err, IsNil)
	c.Assert(index, DeepEquals, expected)
	c.Assert(index, HasLen, 1)
}
