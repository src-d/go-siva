package siva

import (
	"io/ioutil"
	"os"

	. "gopkg.in/check.v1"
)

type ReaderSuite struct{}

var _ = Suite(&ReaderSuite{})

func (s *ReaderSuite) TestIndex(c *C) {
	s.testIndex(c, "fixtures/basic.siva")
}

func (s *ReaderSuite) TestIndexSeveralBlocks(c *C) {
	s.testIndex(c, "fixtures/blocks.siva")
}

func (s *ReaderSuite) testIndex(c *C, fixture string) {
	f, err := os.Open("fixtures/blocks.siva")
	c.Assert(err, IsNil)

	r := NewReader(f)
	i, err := r.Index()
	c.Assert(err, IsNil)
	c.Assert(i, HasLen, 3)
	for j, e := range i {
		c.Assert(e.Name, Equals, files[j].Name)
	}

}

func (s *ReaderSuite) TestGet(c *C) {
	f, err := os.Open("fixtures/blocks.siva")
	c.Assert(err, IsNil)

	r := NewReader(f)
	i, err := r.Index()
	c.Assert(err, IsNil)
	c.Assert(i, HasLen, 3)

	for j, e := range i {
		content, err := r.Get(e)
		c.Assert(err, IsNil)

		bytes, err := ioutil.ReadAll(content)
		c.Assert(err, IsNil)

		c.Assert(string(bytes), Equals, files[j].Body)
	}
}

func (s *ReaderSuite) TestSeekAndRead(c *C) {
	f, err := os.Open("fixtures/blocks.siva")
	c.Assert(err, IsNil)

	r := NewReader(f)
	i, err := r.Index()
	c.Assert(err, IsNil)
	c.Assert(i, HasLen, 3)

	for j, e := range i {
		position, err := r.Seek(e)
		c.Assert(err, IsNil)
		c.Assert(uint64(position), Equals, e.absStart)

		bytes, err := ioutil.ReadAll(r)
		c.Assert(err, IsNil)

		c.Assert(string(bytes), Equals, files[j].Body)
	}
}
