package impl

import "ffsyncclient/cli"

type CLIArgumentsGetRecord struct {
}

func (C CLIArgumentsGetRecord) Mode() cli.Mode {
	return cli.ModeGetRecord
}

func (C CLIArgumentsGetRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsGetRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
