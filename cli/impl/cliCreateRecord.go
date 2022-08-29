package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsCreateRecord struct {
}

func NewCLIArgumentsCreateRecord() *CLIArgumentsCreateRecord {
	return &CLIArgumentsCreateRecord{}
}

func (a *CLIArgumentsCreateRecord) Mode() cli.Mode {
	return cli.ModeCreateRecord
}

func (a *CLIArgumentsCreateRecord) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsCreateRecord) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsCreateRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsCreateRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
