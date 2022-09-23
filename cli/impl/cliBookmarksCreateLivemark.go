package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsBookmarksCreateLivemark struct {
	//TODO

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateLivemark() *CLIArgumentsBookmarksCreateLivemark {
	return &CLIArgumentsBookmarksCreateLivemark{}
}

func (a *CLIArgumentsBookmarksCreateLivemark) Mode() cli.Mode {
	return cli.ModeBookmarksCreateLivemark
}

func (a *CLIArgumentsBookmarksCreateLivemark) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksCreateLivemark) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateLivemark) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateLivemark) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreateLivemark) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
