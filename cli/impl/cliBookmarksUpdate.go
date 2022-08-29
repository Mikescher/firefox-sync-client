package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsBookmarksUpdate struct {
}

func NewCLIArgumentsBookmarksUpdate() *CLIArgumentsBookmarksUpdate {
	return &CLIArgumentsBookmarksUpdate{}
}

func (a *CLIArgumentsBookmarksUpdate) Mode() cli.Mode {
	return cli.ModeBookmarksUpdate
}

func (a *CLIArgumentsBookmarksUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksUpdate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
