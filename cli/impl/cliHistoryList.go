package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsHistoryList struct {
}

func NewCLIArgumentsHistoryList() *CLIArgumentsHistoryList {
	return &CLIArgumentsHistoryList{}
}

func (a *CLIArgumentsHistoryList) Mode() cli.Mode {
	return cli.ModeHistoryList
}

func (a *CLIArgumentsHistoryList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsHistoryList) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryList) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryList) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
