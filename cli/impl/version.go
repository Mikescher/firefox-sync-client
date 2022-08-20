package impl

import (
	"errors"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
)

type CLIArgumentsVersion struct {
}

func (a CLIArgumentsVersion) Mode() cli.Mode {
	return cli.ModeVersion
}

func (a CLIArgumentsVersion) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errors.New("Unknown argument: " + positionalArgs[0])
	}

	if len(optionArgs) > 0 {
		return errors.New("Unknown argument: " + optionArgs[0].Key)
	}

	return nil
}

func (a CLIArgumentsVersion) Execute(ctx *cli.FFSContext) int {
	switch ctx.Opt.Format {
	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutput("{\"version\": \"" + consts.FFSCLIENT_VERSION + "\"}")
		return 0
	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(consts.FFSCLIENT_VERSION)
		return 0
	default:
		ctx.PrintFatalError("Unsupported output-format: " + ctx.Opt.Format.String())
		return 0
	}
}
