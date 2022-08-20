package impl

import "ffsyncclient/cli"

type CLIArgumentsGetRawRecord struct {
}

func (a CLIArgumentsGetRawRecord) Mode() cli.Mode {
	return cli.ModeGetRawRecord
}

func (a CLIArgumentsGetRawRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a CLIArgumentsGetRawRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
