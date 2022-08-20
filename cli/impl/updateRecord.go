package impl

import "ffsyncclient/cli"

type CLIArgumentsUpdateRecord struct {
}

func (C CLIArgumentsUpdateRecord) Mode() cli.Mode {
	return cli.ModeUpdateRecord
}

func (C CLIArgumentsUpdateRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsUpdateRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
