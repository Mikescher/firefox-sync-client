package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsHistoryDelete struct {
}

func NewCLIArgumentsHistoryDelete() *CLIArgumentsHistoryDelete {
	return &CLIArgumentsHistoryDelete{}
}

func (a *CLIArgumentsHistoryDelete) Mode() cli.Mode {
	return cli.ModeHistoryDelete
}

func (a *CLIArgumentsHistoryDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsHistoryDelete) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryDelete) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryDelete) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
