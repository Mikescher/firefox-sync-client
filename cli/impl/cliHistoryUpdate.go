package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
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

func (a *CLIArgumentsHistoryUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsHistoryUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryUpdate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
