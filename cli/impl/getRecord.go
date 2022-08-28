package impl

import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetRecord struct {
}

func NewCLIArgumentsGetRecords() *CLIArgumentsGetRecord {
	return &CLIArgumentsGetRecord{}
}

func (a *CLIArgumentsGetRecord) Mode() cli.Mode {
	return cli.ModeGetRecord
}

func (a *CLIArgumentsGetRecord) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient get <collection> <record-id>", "Get a single record"},
		{"          (--raw | --decoded)", "Return raw data or decoded payload"},
		{"          [--pretty-print]", "Pretty-Print json in decoded data / payload (if possible)"},
	}
}

func (a *CLIArgumentsGetRecord) FullHelp() []string {
	return nil // TODO
}

func (a *CLIArgumentsGetRecord) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetRecord) Execute(ctx *cli.FFSContext) int {
	//TODO implement me
	panic("implement me")
}
