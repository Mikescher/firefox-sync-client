package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsTokenRefresh struct {
}

func NewCLIArgumentsTokenRefresh() *CLIArgumentsTokenRefresh {
	return &CLIArgumentsTokenRefresh{}
}

func (a *CLIArgumentsTokenRefresh) Mode() cli.Mode {
	return cli.ModeTokenRefresh
}

func (a *CLIArgumentsTokenRefresh) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsTokenRefresh) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
