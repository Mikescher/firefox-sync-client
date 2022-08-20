package impl

import "ffsyncclient/cli"

type CLIArgumentsDeleteSingle struct {
}

func (a *CLIArgumentsDeleteSingle) Mode() cli.Mode {
	return cli.ModeDeleteSingle
}

func (a *CLIArgumentsDeleteSingle) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a *CLIArgumentsDeleteSingle) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
