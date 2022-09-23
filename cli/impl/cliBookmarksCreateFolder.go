package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsBookmarksCreateFolder struct {
	//TODO

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateFolder() *CLIArgumentsBookmarksCreateFolder {
	return &CLIArgumentsBookmarksCreateFolder{}
}

func (a *CLIArgumentsBookmarksCreateFolder) Mode() cli.Mode {
	return cli.ModeBookmarksCreateFolder
}

func (a *CLIArgumentsBookmarksCreateFolder) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksCreateFolder) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateFolder) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateFolder) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreateFolder) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
