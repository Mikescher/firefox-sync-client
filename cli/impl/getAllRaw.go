package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetAllRaw struct {
}

func NewCLIArgumentsGetAllRaw() *CLIArgumentsGetAllRaw {
	return &CLIArgumentsGetAllRaw{}
}

func (a *CLIArgumentsGetAllRaw) Mode() cli.Mode {
	return cli.ModeGetAllRaw
}

func (a *CLIArgumentsGetAllRaw) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetAllRaw) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
