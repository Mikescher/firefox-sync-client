package impl

import "ffsyncclient/cli"

type CLIArgumentsHelp struct {
	Extra string
	Verb  *cli.Mode
}

func (C CLIArgumentsHelp) Mode() cli.Mode {
	return cli.ModeHelp
}

func (C CLIArgumentsHelp) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (C CLIArgumentsHelp) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
