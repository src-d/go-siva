package siva

import (
	"bytes"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	if diff := cmp.Diff(expected, footer, cmpopts.IgnoreUnexported(IndexFooter{})); diff != "" {
		c.Fatalf("IndexFooter differs:\n%s", diff)
	}
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
	if diff := cmp.Diff(expected, entry, cmpopts.IgnoreUnexported(IndexEntry{})); diff != "" {
		c.Fatalf("IndexEntry differs:\n%s", diff)
	}
}

func (s *IndexSuite) TestFilter(c *C) {
	i := Index{
		{Header: Header{Name: "foo"}, Start: 1},
		{Header: Header{Name: "foo"}, Start: 2},
	}

	sort.Sort(i)
	f := i.Filter()
	c.Assert(f, HasLen, 1)
	c.Assert(f[0].Start, Equals, uint64(2))
}

func (s *IndexSuite) TestFilterDeleted(c *C) {
	i := Index{
		{Header: Header{Name: "foo"}, Start: 1},
		{Header: Header{Name: "foo", Flags: FlagDeleted}, Start: 2},
	}

	sort.Sort(i)
	f := i.Filter()
	c.Assert(f, HasLen, 0)
}

func (s *IndexSuite) TestFind(c *C) {
	i := Index{
		{Header: Header{Name: "foo"}, Start: 1},
		{Header: Header{Name: "bar"}, Start: 2},
		{Header: Header{Name: "dir/file.txt"}, Start: 3},
	}

	sort.Sort(i)

	e := i.Find("bar")
	c.Assert(e, NotNil)
	c.Assert(e.Start, Equals, uint64(2))

	e = i.Find("dir\\file.txt")
	c.Assert(e, NotNil)
	c.Assert(e.Start, Equals, uint64(3))
}

func (s *IndexSuite) TestToSafePaths(c *C) {
	i := Index{
		{Header: Header{Name: `C:\foo\bar`}, Start: 1},
		{Header: Header{Name: `\\network\share\foo\bar`}, Start: 2},
		{Header: Header{Name: `/foo/bar`}, Start: 3},
		{Header: Header{Name: `../bar`}, Start: 4},
		{Header: Header{Name: `foo/bar/../../baz`}, Start: 5},
	}

	f := i.ToSafePaths()
	expected := Index{
		{Header: Header{Name: `foo/bar`}, Start: 1},
		{Header: Header{Name: `foo/bar`}, Start: 2},
		{Header: Header{Name: `foo/bar`}, Start: 3},
		{Header: Header{Name: `bar`}, Start: 4},
		{Header: Header{Name: `baz`}, Start: 5},
	}
	c.Assert(f, DeepEquals, expected)
}

func (s *IndexSuite) TestGlobPrefix(c *C) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{
			pattern:  "simple",
			expected: "simple",
		},
		{
			pattern:  "*pattern",
			expected: "",
		},
		{
			pattern:  "pattern*",
			expected: "pattern",
		},
		{
			pattern:  "pat*tern*",
			expected: "pat",
		},
		{
			pattern:  "pat\\*tern*",
			expected: "pat\\*tern",
		},
		{
			pattern:  "?pattern",
			expected: "",
		},
		{
			pattern:  "pattern?",
			expected: "pattern",
		},
		{
			pattern:  "pat?tern*",
			expected: "pat",
		},
		{
			pattern:  "pat\\?tern?",
			expected: "pat\\?tern",
		},
		{
			pattern:  "[pattern",
			expected: "",
		},
		{
			pattern:  "pattern[",
			expected: "pattern",
		},
		{
			pattern:  "pat[tern*",
			expected: "pat",
		},
		{
			pattern:  "pat\\[tern[",
			expected: "pat\\[tern",
		},
	}

	for _, test := range tests {
		c.Assert(globPrefix(test.pattern), Equals, test.expected)
	}
}

func BenchmarkGlob5(b *testing.B) {
	benchmarkGlob(0, b)
}
func BenchmarkGlob15(b *testing.B) {
	benchmarkGlob(10, b)
}
func BenchmarkGlob105(b *testing.B) {
	benchmarkGlob(100, b)
}
func BenchmarkGlob10005(b *testing.B) {
	benchmarkGlob(10000, b)
}

func benchmarkGlob(num int, b *testing.B) {
	index := Index{
		{Header: Header{Name: "foo"}, Start: 1},
		{Header: Header{Name: "bar"}, Start: 2},
		{Header: Header{Name: "dir/file.txt"}, Start: 3},
		{Header: Header{Name: "dir/file.png"}, Start: 4},
		{Header: Header{Name: "dir/file.css"}, Start: 5},
	}

	for i := 0; i < num; i++ {
		index = append(index, &IndexEntry{Header: Header{Name: strconv.Itoa(i)}})
	}

	orderedIndex := OrderedIndex(index)
	orderedIndex.Sort()

	globs := []struct {
		name string
		fun  func(string) ([]*IndexEntry, error)
	}{
		{
			name: "normal",
			fun:  index.Glob,
		},
		{
			name: "ordered",
			fun:  orderedIndex.Glob,
		},
	}

	for _, g := range globs {
		b.Run(g.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.fun("dir/*")
			}
		})
	}

}
