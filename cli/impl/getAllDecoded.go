package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetAllDecoded struct {
}

func NewCLIArgumentsGetAllDecoded() *CLIArgumentsGetAllDecoded {
	return &CLIArgumentsGetAllDecoded{}
}

func (a *CLIArgumentsGetAllDecoded) Mode() cli.Mode {
	return cli.ModeGetAllDecoded
}

func (a *CLIArgumentsGetAllDecoded) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetAllDecoded) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
