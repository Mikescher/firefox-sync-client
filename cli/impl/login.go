package impl

import "ffsyncclient/cli"

type CLIArgumentsLogin struct {
}

func (C CLIArgumentsLogin) Mode() cli.Mode {
	return cli.ModeLogin
}

func (C CLIArgumentsLogin) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsLogin) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
