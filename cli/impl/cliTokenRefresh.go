package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
)

type CLIArgumentsTokenRefresh struct {
	Force bool

	CLIArgumentsBaseUtil
}

func NewCLIArgumentsTokenRefresh() *CLIArgumentsTokenRefresh {
	return &CLIArgumentsTokenRefresh{
		Force: false,
	}
}

func (a *CLIArgumentsTokenRefresh) Mode() cli.Mode {
	return cli.ModeTokenRefresh
}

func (a *CLIArgumentsTokenRefresh) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsTokenRefresh) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsTokenRefresh) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient refresh [--force]", "Refresh the current session token (BID Assertion)"},
	}
}

func (a *CLIArgumentsTokenRefresh) FullHelp() []string {
	return []string{
		"$> ffsclient refresh [--force]",
		"",
		"Refresh the current session token",
		"",
		"Use --force to force a new session, even if the old is still valid",
	}
}

func (a *CLIArgumentsTokenRefresh) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "force" && arg.Value == nil {
			a.Force = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsTokenRefresh) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Refresh Token]")
	ctx.PrintVerbose("")

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return err
	}

	if !langext.FileExists(cfp) {
		return fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return err
	}

	ctx.PrintVerbose("Refresh Session Keys")

	session, refreshed, err := client.RefreshSession(ctx, session, a.Force)
	if err != nil {
		return err
	}

	ctx.PrintVerbose("Save session to " + ctx.Opt.SessionFilePath)

	err = session.Save(cfp)
	if err != nil {
		return err
	}

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	if refreshed {
		ctx.PrintPrimaryOutput("Session refreshed")
		return nil
	} else {
		ctx.PrintPrimaryOutput("Session still valid")
		return nil
	}
}
