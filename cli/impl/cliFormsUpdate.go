package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsUpdate struct {
}

func NewCLIArgumentsFormsUpdate() *CLIArgumentsFormsUpdate {
	return &CLIArgumentsFormsUpdate{}
}

func (a *CLIArgumentsFormsUpdate) Mode() cli.Mode {
	return cli.ModeFormsUpdate
}

func (a *CLIArgumentsFormsUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsUpdate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
