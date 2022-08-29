package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsGet struct {
}

func NewCLIArgumentsFormsGet() *CLIArgumentsFormsGet {
	return &CLIArgumentsFormsGet{}
}

func (a *CLIArgumentsFormsGet) Mode() cli.Mode {
	return cli.ModeFormsGet
}

func (a *CLIArgumentsFormsGet) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsGet) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsGet) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
