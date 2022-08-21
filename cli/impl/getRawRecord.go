package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetRawRecord struct {
}

func NewCLIArgumentsGetRawRecord() *CLIArgumentsGetRawRecord {
	return &CLIArgumentsGetRawRecord{}
}

func (a *CLIArgumentsGetRawRecord) Mode() cli.Mode {
	return cli.ModeGetRawRecord
}

func (a *CLIArgumentsGetRawRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetRawRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
