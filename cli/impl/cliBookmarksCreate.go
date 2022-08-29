package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsBookmarksCreate struct {
}

func NewCLIArgumentsBookmarksCreate() *CLIArgumentsBookmarksCreate {
	return &CLIArgumentsBookmarksCreate{}
}

func (a *CLIArgumentsBookmarksCreate) Mode() cli.Mode {
	return cli.ModeBookmarksCreate
}

func (a *CLIArgumentsBookmarksCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsBookmarksCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
