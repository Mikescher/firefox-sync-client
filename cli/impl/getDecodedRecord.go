package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetDecodedRecord struct {
}

func NewCLIArgumentsGetDecodedRecord() *CLIArgumentsGetDecodedRecord {
	return &CLIArgumentsGetDecodedRecord{}
}

func (a *CLIArgumentsGetDecodedRecord) Mode() cli.Mode {
	return cli.ModeGetDecodedRecord
}

func (a *CLIArgumentsGetDecodedRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetDecodedRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
