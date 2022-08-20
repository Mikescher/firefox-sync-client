package impl

import "ffsyncclient/cli"

type CLIArgumentsDeleteAll struct {
}

func (a *CLIArgumentsDeleteAll) Mode() cli.Mode {
	return cli.ModeDeleteAll
}

func (a *CLIArgumentsDeleteAll) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a *CLIArgumentsDeleteAll) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
