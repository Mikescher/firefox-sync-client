package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsDeleteAll struct {
}

func NewCLIArgumentsDeleteAll() *CLIArgumentsDeleteAll {
	return &CLIArgumentsDeleteAll{}
}

func (a *CLIArgumentsDeleteAll) Mode() cli.Mode {
	return cli.ModeDeleteAll
}

func (a *CLIArgumentsDeleteAll) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsDeleteAll) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
