package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsDelete struct {
}

func NewCLIArgumentsFormsDelete() *CLIArgumentsFormsDelete {
	return &CLIArgumentsFormsDelete{}
}

func (a *CLIArgumentsFormsDelete) Mode() cli.Mode {
	return cli.ModeFormsDelete
}

func (a *CLIArgumentsFormsDelete) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsDelete) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsDelete) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
