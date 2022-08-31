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

func (a *CLIArgumentsTokenRefresh) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Refresh Token]")
	ctx.PrintVerbose("")

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

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Refresh Session Keys")

	session, refreshed, err := client.RefreshSession(ctx, session, a.Force)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Save session to " + ctx.Opt.SessionFilePath)

	err = session.Save(cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	if refreshed {
		ctx.PrintPrimaryOutput("Session refreshed")
		return 0
	} else {
		ctx.PrintPrimaryOutput("Session still valid")
		return 0
	}
}
