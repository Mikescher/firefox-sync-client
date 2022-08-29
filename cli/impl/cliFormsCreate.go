package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsCreate struct {
}

func NewCLIArgumentsFormsCreate() *CLIArgumentsFormsCreate {
	return &CLIArgumentsFormsCreate{}
}

func (a *CLIArgumentsFormsCreate) Mode() cli.Mode {
	return cli.ModeFormsCreate
}

func (a *CLIArgumentsFormsCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsFormsCreate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsCreate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsCreate) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
