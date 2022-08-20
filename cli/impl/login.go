package impl

import "ffsyncclient/cli"

type CLIArgumentsLogin struct {
}

func (a CLIArgumentsLogin) Mode() cli.Mode {
	return cli.ModeLogin
}

func (a CLIArgumentsLogin) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a CLIArgumentsLogin) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
