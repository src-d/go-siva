package main

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type CmdList struct {
	cmd
}

func (c *CmdList) Execute(args []string) error {
	if err := c.buildReader(); err != nil {
		return err
	}

	defer c.close()
	if err := c.listVolume(); err != nil {
		return err
	}

	return nil
}

func (c *CmdList) listVolume() error {
	i, err := c.r.Index()
	if err != nil {
		return fmt.Errorf("error reading index: %s", err)
	}

	for _, file := range i {
		fmt.Printf("%s %s % 6s %s\n",
			file.Mode.Perm(),
			file.ModTime.Format("Jan 02 15:04"),
			humanize.Bytes(file.Size),
			file.Name,
		)
	}

	return nil
}
