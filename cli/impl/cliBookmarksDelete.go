package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsBookmarksDelete struct {
}

func NewCLIArgumentsBookmarksDelete() *CLIArgumentsBookmarksDelete {
	return &CLIArgumentsBookmarksDelete{}
}

func (a *CLIArgumentsBookmarksDelete) Mode() cli.Mode {
	return cli.ModeBookmarksDelete
}

func (a *CLIArgumentsBookmarksDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksDelete) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksDelete) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksDelete) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
