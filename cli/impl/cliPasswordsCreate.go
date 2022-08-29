package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsPasswordsCreate struct {
}

func NewCLIArgumentsPasswordsCreate() *CLIArgumentsPasswordsCreate {
	return &CLIArgumentsPasswordsCreate{}
}

func (a *CLIArgumentsPasswordsCreate) Mode() cli.Mode {
	return cli.ModePasswordsCreate
}

func (a *CLIArgumentsPasswordsCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
