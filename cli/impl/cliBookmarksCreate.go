package impl

import (
	"ffsyncclient/cli"
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

func (a *CLIArgumentsBookmarksCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
