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
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
	}

	buf := new(bytes.Buffer)
	tw := NewWriter(buf)
	for _, file := range files {
		hdr := &Header{
			Name:    file.Name,
			Mode:    0600,
			ModTime: time.Now(),
		}

		err := tw.WriteHeader(hdr)
		c.Assert(err, IsNil)

		n, err := tw.Write([]byte(file.Body))
		c.Assert(err, IsNil)
		c.Assert(n, Equals, len(file.Body))
	}

	err := tw.Close()
	c.Assert(err, IsNil)

	r := bytes.NewReader(buf.Bytes())
	tr := NewReader(r)

	index, err := tr.Index()
	c.Assert(err, IsNil)

	// Iterate through the files in the archive.
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
