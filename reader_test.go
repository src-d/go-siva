package siva

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

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

func (s *ReaderSuite) TestIndexOverwrittenFiles(c *C) {
	s.testIndex(c, "fixtures/overwritten.siva")
}

func (s *ReaderSuite) testIndex(c *C, fixture string) {
	f, err := os.Open(fixture)
	c.Assert(err, IsNil)

	r := NewReader(f)
	i, err := r.Index()
	i = i.Filter()
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

func (s *ReaderSuite) TestIndexGlob(c *C) {
	s.testIndexGlob(c, false)
}

func (s *ReaderSuite) TestIndexGlobOrdered(c *C) {
	s.testIndexGlob(c, true)
}

func (s *ReaderSuite) testIndexGlob(c *C, ordered bool) {
	s.testIndexGlobSingle(c, "*", ordered, []string{
		"file.txt",
	})
	s.testIndexGlobSingle(c, "*/*", ordered, []string{
		"letters/a",
		"letters/b",
		"letters/c",
		"numbers/1",
		"numbers/2",
		"numbers/3",
	})
	s.testIndexGlobSingle(c, "letters/*", ordered, []string{
		"letters/a",
		"letters/b",
		"letters/c",
	})
	s.testIndexGlobSingle(c, "numbers\\*", ordered, []string{
		"numbers/1",
		"numbers/2",
		"numbers/3",
	})
	s.testIndexGlobSingle(c, "nonexistent/*", ordered, []string{})
}

func (s *ReaderSuite) testIndexGlobSingle(
	c *C,
	pattern string,
	ordered bool,
	expected []string,
) {
	f, err := os.Open("fixtures/dirs.siva")
	c.Assert(err, IsNil)

	r := NewReader(f)
	i, err := r.Index()
	c.Assert(err, IsNil)
	c.Assert(i, HasLen, 7)

	var matches []*IndexEntry
	if ordered {
		o := OrderedIndex(i)
		o.Sort()
		matches, err = o.Glob(pattern)
	} else {
		matches, err = i.Glob(pattern)
	}
	c.Assert(err, IsNil)
	matchNames := []string{}
	for _, match := range matches {
		matchNames = append(matchNames, match.Name)
	}
	sort.Strings(matchNames)

	c.Assert(matchNames, DeepEquals, expected)

	c.Assert(f.Close(), IsNil)
}

func (s *ReaderSuite) TestOffset(c *C) {
	data, err := ioutil.ReadFile("fixtures/basic.siva")
	c.Assert(err, IsNil)

	indexOffset := uint64(len(data))
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)
	buf := bytes.NewReader(data)

	r := NewReader(buf)
	_, err = r.Index()
	_, ok := err.(*IndexReadError)
	c.Assert(ok, Equals, true)

	r = NewReaderWithOffset(buf, indexOffset)
	i, err := r.Index()
	c.Assert(err, IsNil)

	entry := i.Find("gopher.txt")
	c.Assert(entry, NotNil)
	c.Assert(entry.Size, Equals, uint64(35))
}

func (s *IndexSuite) TestUnsupportedIndexVersion(c *C) {
	f, err := os.Open(filepath.Join("fixtures", "basicv2.siva"))
	c.Assert(err, IsNil)

	r := NewReader(f)
	_, err = r.Index()
	_, ok := err.(*IndexReadError)
	c.Assert(ok, Equals, true)
}
