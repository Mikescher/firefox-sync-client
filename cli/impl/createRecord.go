package impl

import "ffsyncclient/cli"

type CLIArgumentsCreateRecord struct {
}

func (C CLIArgumentsCreateRecord) Mode() cli.Mode {
	return cli.ModeCreateRecord
}

func (C CLIArgumentsCreateRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsCreateRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
