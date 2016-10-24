package siva_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-siva.v1"
)

type ReadWriterSuite struct {
	tmpDir string
}

var _ = Suite(&ReadWriterSuite{})

func (s *ReadWriterSuite) SetUpSuite(c *C) {
	s.tmpDir = c.MkDir()
}

func (s *ReadWriterSuite) TestWriteRead(c *C) {
	tmpFile, err := os.Create(filepath.Join(s.tmpDir, c.TestName()))
	c.Assert(err, IsNil)
	c.Assert(tmpFile, NotNil)

	rw, err := siva.NewReaderWriter(tmpFile)
	c.Assert(err, IsNil)
	c.Assert(rw, NotNil)

	for i := 0; i < 100; i++ {
		curName := fmt.Sprintf("foo-%d", i)
		content := strings.Repeat("#", i)

		err := rw.WriteHeader(&siva.Header{
			Name: curName,
		})
		c.Assert(err, IsNil)

		written, err := rw.Write([]byte(content))
		c.Assert(err, IsNil)
		c.Assert(written, Equals, i)

		err = rw.Flush()
		c.Assert(err, IsNil)

		index, err := rw.Index()
		c.Assert(err, IsNil)
		c.Assert(len(index), Equals, i+1)

		e := index.Find(curName)
		c.Assert(e, NotNil)

		sr, err := rw.Get(e)
		c.Assert(err, IsNil)
		c.Assert(sr, NotNil)

		read, err := ioutil.ReadAll(sr)
		c.Assert(err, IsNil)
		c.Assert(string(read), Equals, content)
	}

	c.Assert(rw.Close(), IsNil)
}

func (s *ReadWriterSuite) TestReadExisting(c *C) {
	f, err := os.OpenFile("fixtures/basic.siva", os.O_RDONLY, os.ModePerm)
	c.Assert(err, IsNil)
	c.Assert(f, NotNil)

	rw, err := siva.NewReaderWriter(f)
	c.Assert(err, IsNil)
	c.Assert(rw, NotNil)

	index, err := rw.Index()
	c.Assert(err, IsNil)
	c.Assert(len(index), Equals, 3)

	c.Assert(rw.Close(), IsNil)
}

func (s *ReadWriterSuite) TestFailIfNotReadAt(c *C) {
	rw, err := siva.NewReaderWriter(&dummyReadWriterSeeker{})
	c.Assert(err, Equals, siva.ErrInvalidReaderAt)
	c.Assert(rw, IsNil)
}

type dummyReadWriterSeeker struct {
}

func (_ dummyReadWriterSeeker) Read(p []byte) (n int, err error) {
	return
}

func (_ dummyReadWriterSeeker) Write(p []byte) (n int, err error) {
	return
}

func (_ dummyReadWriterSeeker) Seek(offset int64, whence int) (n int64, err error) {
	return
}
