package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsRecordsUpdate struct {
}

func NewCLIArgumentsRecordsUpdate() *CLIArgumentsRecordsUpdate {
	return &CLIArgumentsRecordsUpdate{}
}

func (a *CLIArgumentsRecordsUpdate) Mode() cli.Mode {
	return cli.ModeRecordsUpdate
}

func (a *CLIArgumentsRecordsUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0) //TODO
}

func (a *CLIArgumentsRecordsUpdate) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsRecordsUpdate) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsRecordsUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsUpdate) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
