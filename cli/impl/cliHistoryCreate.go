package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsHistoryCreate struct {
}

func NewCLIArgumentsHistoryCreate() *CLIArgumentsHistoryCreate {
	return &CLIArgumentsHistoryCreate{}
}

func (a *CLIArgumentsHistoryCreate) Mode() cli.Mode {
	return cli.ModeHistoryCreate
}

func (a *CLIArgumentsHistoryCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
