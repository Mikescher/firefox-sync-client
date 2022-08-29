package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsPasswordsGet struct {
}

func NewCLIArgumentsPasswordsGet() *CLIArgumentsPasswordsGet {
	return &CLIArgumentsPasswordsGet{}
}

func (a *CLIArgumentsPasswordsGet) Mode() cli.Mode {
	return cli.ModePasswordsGet
}

func (a *CLIArgumentsPasswordsGet) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsGet) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsGet) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
