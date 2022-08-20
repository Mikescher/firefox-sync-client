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
	_, err := os.Stdout.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stdout")
	}
}

func (c FFSContext) PrintFatalError(msg string) {
	_, err := os.Stderr.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stderr")
	}
}

func NewContext(opt Options) *FFSContext {
	return &FFSContext{
		Context: context.Background(),
		Opt:     opt,
	}
}
