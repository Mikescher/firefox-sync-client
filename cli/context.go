package cli

import (
	"context"
	"os"
)

type FFSContext struct {
	context.Context
	Opt Options
}

func (c FFSContext) PrintPrimaryOutput(msg string) {
	if c.Opt.Quiet {
		return
	}

	_, err := os.Stdout.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stdout: " + err.Error())
	}
}

func (c FFSContext) PrintFatalError(msg string) {
	if c.Opt.Quiet {
		return
	}

	_, err := os.Stderr.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stderr: " + err.Error())
	}
}

func NewContext(opt Options) *FFSContext {
	return &FFSContext{
		Context: context.Background(),
		Opt:     opt,
	}
}
