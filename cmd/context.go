package cmd

import (
	"io"
)

type CommandContext struct {
	OneBlock *OneBlockOptions
	StdOut   io.Writer
	StdErr   io.Writer
	StdIn    io.Reader
}
