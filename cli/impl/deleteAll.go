package impl

import "ffsyncclient/cli"

type CLIArgumentsDeleteAll struct {
}

func (C CLIArgumentsDeleteAll) Mode() cli.Mode {
	return cli.ModeDeleteAll
}

func (C CLIArgumentsDeleteAll) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsDeleteAll) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
