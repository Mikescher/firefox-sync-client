package impl

import "ffsyncclient/cli"

type CLIArgumentsUpdateRecord struct {
}

func (a CLIArgumentsUpdateRecord) Mode() cli.Mode {
	return cli.ModeUpdateRecord
}

func (a CLIArgumentsUpdateRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	//TODO implement me
	panic("implement me")
}

func (a CLIArgumentsUpdateRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
