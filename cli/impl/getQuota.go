package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetQuota struct {
}

func NewCLIArgumentsGetQuota() *CLIArgumentsGetQuota {
	return &CLIArgumentsGetQuota{}
}

func (a *CLIArgumentsGetQuota) Mode() cli.Mode {
	return cli.ModeGetQuota
}

func (a *CLIArgumentsGetQuota) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetQuota) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
