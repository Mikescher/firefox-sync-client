package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsBookmarksDelete struct {
}

func NewCLIArgumentsBookmarksDelete() *CLIArgumentsBookmarksDelete {
	return &CLIArgumentsBookmarksDelete{}
}

func (a *CLIArgumentsBookmarksDelete) Mode() cli.Mode {
	return cli.ModeBookmarksDelete
}

func (a *CLIArgumentsBookmarksDelete) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksDelete) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsBookmarksDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksDelete) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
