package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsHistoryUpdate struct {
}

func NewCLIArgumentsHistoryUpdate() *CLIArgumentsHistoryUpdate {
	return &CLIArgumentsHistoryUpdate{}
}

func (a *CLIArgumentsHistoryUpdate) Mode() cli.Mode {
	return cli.ModeHistoryUpdate
}

func (a *CLIArgumentsHistoryUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryUpdate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
