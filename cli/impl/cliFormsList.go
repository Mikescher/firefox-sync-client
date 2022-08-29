package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsList struct {
}

func NewCLIArgumentsFormsList() *CLIArgumentsFormsList {
	return &CLIArgumentsFormsList{}
}

func (a *CLIArgumentsFormsList) Mode() cli.Mode {
	return cli.ModeFormsList
}

func (a *CLIArgumentsFormsList) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsList) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsList) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
