package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/src-d/siva"
)

type CmdPack struct {
	cmd
	Verbose bool `short:"v" description:"Activates the verbose mode"`
	Append  bool `long:"append" description:"If append, the files are added to an existing siva file"`
	Input   struct {
		Files []string `positional-arg-name:"input" description:"files or directories to be add to the archive."`
	} `positional-args:"yes"`
}

func (c *CmdPack) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.do(); err != nil {
		if err := os.Remove(c.Args.File); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (c *CmdPack) do() error {
	if err := c.buildWriter(c.Append); err != nil {
		return err
	}

	defer c.close()
	if err := c.pack(); err != nil {
		return err
	}

	return nil
}

func (c *CmdPack) validate() error {
	if err := c.cmd.validate(); err != nil {
		return err
	}

	if len(c.Input.Files) == 0 {
		return fmt.Errorf("Invalid input count, please add one or more input files/dirs")
	}

	return nil
}

func (c *CmdPack) pack() error {
	for _, file := range c.Input.Files {
		fi, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("Invalid input file/dir %q, no such file", file)
		}

		if err := c.packFile(file, fi); err != nil {
			return err
		}
	}

	return nil
}

func (c *CmdPack) packFile(fullpath string, fi os.FileInfo) error {
	if !fi.Mode().IsRegular() {
		return nil
	}

	if c.Verbose {
		fmt.Println(fullpath)
	}

	h := &siva.Header{
		Name:    cleanPath(fullpath),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}

	if err := c.w.WriteHeader(h); err != nil {
		return err
	}

	f, err := os.Open(fullpath)
	if err != nil {
		return err
	}

	n, err := io.Copy(c.w, f)
	if err != nil {
		return err
	}

	if n != fi.Size() {
		return fmt.Errorf("unexpected bytes written")
	}

	return nil
}

func cleanPath(path string) string {
	path = filepath.Clean(path)
	for len(path) >= 3 && path[:3] == "../" {
		path = path[3:]
	}

	for len(path) >= 2 && path[:2] == "./" {
		path = path[2:]
	}

	if len(path) > 1 && path[:1] == "/" {
		path = path[1:]
	}

	return path
}
