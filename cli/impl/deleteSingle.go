package impl

import "ffsyncclient/cli"

type CLIArgumentsDeleteSingle struct {
}

func (C CLIArgumentsDeleteSingle) Mode() cli.Mode {
	return cli.ModeDeleteSingle
}

func (C CLIArgumentsDeleteSingle) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsDeleteSingle) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
