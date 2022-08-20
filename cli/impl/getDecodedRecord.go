package impl

import "ffsyncclient/cli"

type CLIArgumentsGetDecodedRecord struct {
}

func (a *CLIArgumentsGetDecodedRecord) Mode() cli.Mode {
	return cli.ModeGetDecodedRecord
}

func (a *CLIArgumentsGetDecodedRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a *CLIArgumentsGetDecodedRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
