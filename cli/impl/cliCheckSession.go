package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsCheckSession struct {
}

func NewCLIArgumentsCheckSession() *CLIArgumentsCheckSession {
	return &CLIArgumentsCheckSession{}
}

func (a *CLIArgumentsCheckSession) Mode() cli.Mode {
	return cli.ModeCheckSession
}

func (a *CLIArgumentsCheckSession) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient check-session", "Verify that the current session is valid"},
	}
}

func (a *CLIArgumentsCheckSession) FullHelp() []string {
	return []string{
		"$> ffsclient check-session",
		"",
		"Validate that the current session is valid.",
		"",
		"This does not mean that browser-id-assertion is valid.",
		"The BID can still be expired and would need to be refreshed.",
		"(see `ffsclient refresh --help`)",
	}
}

func (a *CLIArgumentsCheckSession) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsCheckSession) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Check Session]")
	ctx.PrintVerbose("")

	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !langext.FileExists(cfp) {
		ctx.PrintFatalMessage("Sessionfile does not exist.")
		ctx.PrintFatalMessage("Use `ffsclient login <email> <password>` first")
		return consts.ExitcodeNoLogin
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	okay, err := client.CheckSession(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	return a.printOutput(ctx, okay)
}

func (a *CLIArgumentsCheckSession) printOutput(ctx *cli.FFSContext, okay bool) int {

	ec := 0
	if !okay {
		ec = consts.ExitcodeInvalidSession
	}

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		if okay {
			ctx.PrintPrimaryOutput("Okay")
		} else {
			ctx.PrintPrimaryOutput("Session invalid")
		}
		return ec

	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(langext.H{"valid": okay})
		return ec

	case cli.OutputFormatXML:
		type xml struct {
			Valid   bool     `xml:",innerxml"`
			XMLName struct{} `xml:"Valid"`
		}
		ctx.PrintPrimaryOutputXML(xml{Valid: okay})
		return ec

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}
