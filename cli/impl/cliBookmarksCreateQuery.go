package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsBookmarksCreateQuery struct {
	//TODO

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateQuery() *CLIArgumentsBookmarksCreateQuery {
	return &CLIArgumentsBookmarksCreateQuery{}
}

func (a *CLIArgumentsBookmarksCreateQuery) Mode() cli.Mode {
	return cli.ModeBookmarksCreateQuery
}

func (a *CLIArgumentsBookmarksCreateQuery) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksCreateQuery) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateQuery) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateQuery) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreateQuery) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
