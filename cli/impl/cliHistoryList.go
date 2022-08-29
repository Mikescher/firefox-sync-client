package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsHistoryList struct {
}

func NewCLIArgumentsHistoryList() *CLIArgumentsHistoryList {
	return &CLIArgumentsHistoryList{}
}

func (a *CLIArgumentsHistoryList) Mode() cli.Mode {
	return cli.ModeHistoryList
}

func (a *CLIArgumentsHistoryList) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryList) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsHistoryList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryList) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
