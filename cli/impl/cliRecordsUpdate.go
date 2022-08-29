package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsUpdateRecord struct {
}

func NewCLIArgumentsRecordsUpdate() *CLIArgumentsUpdateRecord {
	return &CLIArgumentsUpdateRecord{}
}

func (a *CLIArgumentsUpdateRecord) Mode() cli.Mode {
	return cli.ModeRecordsUpdate
}

func (a *CLIArgumentsUpdateRecord) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsUpdateRecord) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsUpdateRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsUpdateRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
