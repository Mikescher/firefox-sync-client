package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetRecord struct {
}

func NewCLIArgumentsGetRecords() *CLIArgumentsGetRecord {
	return &CLIArgumentsGetRecord{}
}

func (a *CLIArgumentsGetRecord) Mode() cli.Mode {
	return cli.ModeGetRecord
}

func (a *CLIArgumentsGetRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
