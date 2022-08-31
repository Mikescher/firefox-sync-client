package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
)

type CLIArgumentsFormsList struct {
}

func NewCLIArgumentsFormsList() *CLIArgumentsFormsList {
	return &CLIArgumentsFormsList{}
}

func (a *CLIArgumentsFormsList) Mode() cli.Mode {
	return cli.ModeFormsList
}

func (a *CLIArgumentsFormsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsFormsList) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsFormsList) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsFormsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsList) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
