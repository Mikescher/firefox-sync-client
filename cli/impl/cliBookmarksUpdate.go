package impl

import (
	"ffsyncclient/cli"
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

func (a *CLIArgumentsBookmarksUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksUpdate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
