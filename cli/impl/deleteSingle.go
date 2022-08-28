package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsDeleteSingle struct {
}

func NewCLIArgumentsDeleteSingle() *CLIArgumentsDeleteSingle {
	return &CLIArgumentsDeleteSingle{}
}

func (a *CLIArgumentsDeleteSingle) Mode() cli.Mode {
	return cli.ModeDeleteSingle
}

func (a *CLIArgumentsDeleteSingle) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient delete <record-id>", "Delete the specified record"},
	}
}

func (a *CLIArgumentsDeleteSingle) FullHelp() []string {
	return nil // TODO
}

func (a *CLIArgumentsDeleteSingle) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsDeleteSingle) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
