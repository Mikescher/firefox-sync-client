package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
)

type CLIArgumentsVersion struct {
}

func NewCLIArgumentsVersion() *CLIArgumentsVersion {
	return &CLIArgumentsVersion{}
}

func (a *CLIArgumentsVersion) Mode() cli.Mode {
	return cli.ModeVersion
}

func (a *CLIArgumentsVersion) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsVersion) Execute(ctx *cli.FFSContext) int {
	type xml struct {
		Version string   `xml:"Version,attr"`
		XMLName struct{} `xml:"FirefoxSyncClient"`
	}

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {
	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(consts.FFSCLIENT_VERSION)
		return 0
	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(langext.H{"version": consts.FFSCLIENT_VERSION})
		return 0
	case cli.OutputFormatXML:
		ctx.PrintPrimaryOutputXML(xml{Version: consts.FFSCLIENT_VERSION})
		return 0
	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return 0
	}
}
