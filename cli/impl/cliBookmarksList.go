package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsBookmarksList struct {
}

func NewCLIArgumentsBookmarksList() *CLIArgumentsBookmarksList {
	return &CLIArgumentsBookmarksList{}
}

func (a *CLIArgumentsBookmarksList) Mode() cli.Mode {
	return cli.ModeBookmarksList
}

func (a *CLIArgumentsBookmarksList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksList) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksList) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksList) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
