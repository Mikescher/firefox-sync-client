package impl

import "ffsyncclient/cli"

type CLIArgumentsListCollections struct {
}

func (a *CLIArgumentsListCollections) Mode() cli.Mode {
	return cli.ModeListCollections
}

func (a *CLIArgumentsListCollections) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a *CLIArgumentsListCollections) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
