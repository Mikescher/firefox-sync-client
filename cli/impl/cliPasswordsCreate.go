package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
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

func (a *CLIArgumentsPasswordsCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsPasswordsCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
