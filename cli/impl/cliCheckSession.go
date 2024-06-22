package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/syncclient"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsCheckSession struct {
	CLIArgumentsBaseUtil
}

func NewCLIArgumentsCheckSession() *CLIArgumentsCheckSession {
	return &CLIArgumentsCheckSession{}
}

func (a *CLIArgumentsCheckSession) Mode() cli.Mode {
	return cli.ModeCheckSession
}

func (a *CLIArgumentsCheckSession) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsCheckSession) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
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
		"This does not mean that AccessToken is valid.",
		"The OAuth can still be expired and would need to be refreshed.",
		"(see `ffsclient refresh --help`)",
	}
}

func (a *CLIArgumentsCheckSession) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsCheckSession) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Check Session]")
	ctx.PrintVerbose("")

	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return err
	}

	if !langext.FileExists(cfp) {
		return fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx, ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return err
	}

	// ========================================================================

	okay, err := client.CheckSession(ctx, session)
	if err != nil {
		return err
	}

	// ========================================================================

	return a.printOutput(ctx, okay)
}

func (a *CLIArgumentsCheckSession) printOutput(ctx *cli.FFSContext, okay bool) error {

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		if okay {
			ctx.PrintPrimaryOutput("Okay")
			return nil
		} else {
			ctx.PrintPrimaryOutput("Session invalid")
			return fferr.NewEmpty(consts.ExitcodeInvalidSession)
		}

	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(langext.H{"valid": okay})
		if okay {
			return nil
		} else {
			return fferr.NewEmpty(consts.ExitcodeInvalidSession)
		}

	case cli.OutputFormatXML:
		type xml struct {
			Valid   bool     `xml:",chardata"`
			XMLName struct{} `xml:"Valid"`
		}
		ctx.PrintPrimaryOutputXML(xml{Valid: okay})
		if okay {
			return nil
		} else {
			return fferr.NewEmpty(consts.ExitcodeInvalidSession)
		}

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
