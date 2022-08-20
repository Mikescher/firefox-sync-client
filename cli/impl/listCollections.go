package impl

import "ffsyncclient/cli"

type CLIArgumentsListCollections struct {
}

func (C CLIArgumentsListCollections) Mode() cli.Mode {
	return cli.ModeListCollections
}

func (C CLIArgumentsListCollections) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsListCollections) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
