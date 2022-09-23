package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsBookmarksCreateSeparator struct {
	//TODO

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateSeparator() *CLIArgumentsBookmarksCreateSeparator {
	return &CLIArgumentsBookmarksCreateSeparator{}
}

func (a *CLIArgumentsBookmarksCreateSeparator) Mode() cli.Mode {
	return cli.ModeBookmarksCreateSeparator
}

func (a *CLIArgumentsBookmarksCreateSeparator) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksCreateSeparator) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateSeparator) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreateSeparator) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreateSeparator) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
