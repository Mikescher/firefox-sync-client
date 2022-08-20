package impl

import "ffsyncclient/cli"

type CLIArgumentsCreateRecord struct {
}

func (a CLIArgumentsCreateRecord) Mode() cli.Mode {
	return cli.ModeCreateRecord
}

func (a CLIArgumentsCreateRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a CLIArgumentsCreateRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
