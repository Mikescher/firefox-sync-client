package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsBookmarksList struct {
}

func NewCLIArgumentsBookmarksList() *CLIArgumentsBookmarksList {
	return &CLIArgumentsBookmarksList{}
}

func (a *CLIArgumentsBookmarksList) Mode() cli.Mode {
	return cli.ModeBookmarksList
}

func (a *CLIArgumentsBookmarksList) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksList) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksList) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
