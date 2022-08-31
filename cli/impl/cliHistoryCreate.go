package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsHistoryCreate struct {
}

func NewCLIArgumentsHistoryCreate() *CLIArgumentsHistoryCreate {
	return &CLIArgumentsHistoryCreate{}
}

func (a *CLIArgumentsHistoryCreate) Mode() cli.Mode {
	return cli.ModeHistoryCreate
}

func (a *CLIArgumentsHistoryCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsHistoryCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
