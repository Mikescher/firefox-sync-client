package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsFormsDelete struct {
}

func NewCLIArgumentsFormsDelete() *CLIArgumentsFormsDelete {
	return &CLIArgumentsFormsDelete{}
}

func (a *CLIArgumentsFormsDelete) Mode() cli.Mode {
	return cli.ModeFormsDelete
}

func (a *CLIArgumentsFormsDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsFormsDelete) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsDelete) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsDelete) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
