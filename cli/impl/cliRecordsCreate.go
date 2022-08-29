package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsRecordsCreate struct {
}

func NewCLIArgumentsRecordsCreate() *CLIArgumentsRecordsCreate {
	return &CLIArgumentsRecordsCreate{}
}

func (a *CLIArgumentsRecordsCreate) Mode() cli.Mode {
	return cli.ModeRecordsCreate
}

func (a *CLIArgumentsRecordsCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsRecordsCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsRecordsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsCreate) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
