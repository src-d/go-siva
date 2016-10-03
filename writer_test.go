package iba

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	. "gopkg.in/check.v1"
)

type WriterSuite struct{}

var _ = Suite(&WriterSuite{})

func (s *WriterSuite) TestWriterReaderIdempotent(c *C) {
	buf := new(bytes.Buffer)
	tw := NewWriter(buf, 0)
	for _, file := range files {
		s.writeFixture(c, tw, file)
	}

	err := tw.Close()
	c.Assert(err, IsNil)

	r := bytes.NewReader(buf.Bytes())
	tr := NewReader(r)

	index, err := tr.Index()
	c.Assert(err, IsNil)
	c.Assert(index, HasLen, 3)

	for i, e := range index {
		_, err := tr.Seek(e)
		if err == io.EOF {
			break
		}

		c.Assert(err, IsNil)
		c.Assert(e.Name, Equals, files[i].Name)

		content, err := ioutil.ReadAll(tr)
		c.Assert(err, IsNil)
		c.Assert(string(content), Equals, files[i].Body)
	}
}

func (s *WriterSuite) TestWriterReaderIdempotentMultiWrite(c *C) {
	buf := new(bytes.Buffer)
	tw := NewWriter(buf, 0)
	for _, file := range files[1:] {
		s.writeFixture(c, tw, file)
	}

	err := tw.Close()
	c.Assert(err, IsNil)

	tw = NewWriter(buf, uint64(buf.Len()))
	for _, file := range files[0:1] {
		s.writeFixture(c, tw, file)
	}

	err = tw.Close()
	c.Assert(err, IsNil)

	r := bytes.NewReader(buf.Bytes())
	tr := NewReader(r)

	index, err := tr.Index()
	c.Assert(err, IsNil)
	c.Assert(index, HasLen, 3)

	for i, e := range index {
		_, err := tr.Seek(e)
		if err == io.EOF {
			break
		}

		c.Assert(err, IsNil)
		c.Assert(e.Name, Equals, files[i].Name)

		content, err := ioutil.ReadAll(tr)
		c.Assert(err, IsNil)
		c.Assert(string(content), Equals, files[i].Body)
	}
}

func (s *WriterSuite) writeFixture(c *C, w *Writer, file fileFixture) {
	hdr := &Header{
		Name:    file.Name,
		Mode:    0600,
		ModTime: time.Now(),
	}

	err := w.WriteHeader(hdr)
	c.Assert(err, IsNil)

	n, err := w.Write([]byte(file.Body))
	c.Assert(err, IsNil)
	c.Assert(n, Equals, len(file.Body))
}

type fileFixture struct {
	Name, Body string
}

var files = []fileFixture{
	{"readme.txt", "This archive contains some text files."},
	{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"todo.txt", "Get animal handling license."},
}
