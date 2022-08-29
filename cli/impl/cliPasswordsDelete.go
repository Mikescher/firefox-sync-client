package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsPasswordsDelete struct {
}

func NewCLIArgumentsPasswordsDelete() *CLIArgumentsPasswordsDelete {
	return &CLIArgumentsPasswordsDelete{}
}

func (a *CLIArgumentsPasswordsDelete) Mode() cli.Mode {
	return cli.ModePasswordsDelete
}

func (a *CLIArgumentsPasswordsDelete) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsDelete) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsDelete) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
