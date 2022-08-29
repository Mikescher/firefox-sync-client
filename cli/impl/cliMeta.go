package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsMeta struct {
}

func NewCLIArgumentsMeta() *CLIArgumentsMeta {
	return &CLIArgumentsMeta{}
}

func (a *CLIArgumentsMeta) Mode() cli.Mode {
	return cli.ModeMeta
}

func (a *CLIArgumentsMeta) ShortHelp() [][]string {
	return nil //TODO
}

func (a *CLIArgumentsMeta) FullHelp() []string {
	return nil //TODO
}

func (a *CLIArgumentsMeta) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsMeta) Execute(ctx *cli.FFSContext) int {
	panic("implement me") //TODO implement me
}
