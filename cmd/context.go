package cmd

import (
	"io"
)

type CommandContext struct {
	*Oneblock
	StdOut io.Writer
	StdErr io.Writer
	StdIn  io.Reader
}
