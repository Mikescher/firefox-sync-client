package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsPasswordsUpdate struct {
}

func NewCLIArgumentsPasswordsUpdate() *CLIArgumentsPasswordsUpdate {
	return &CLIArgumentsPasswordsUpdate{}
}

func (a *CLIArgumentsPasswordsUpdate) Mode() cli.Mode {
	return cli.ModePasswordsUpdate
}

func (a *CLIArgumentsPasswordsUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsPasswordsUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsPasswordsUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsUpdate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
