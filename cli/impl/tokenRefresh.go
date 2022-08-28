package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
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
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		if arg.Key == "force" && arg.Value == nil {
			a.Force = true
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
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

	_, refreshed, err := client.RefreshSession(ctx, session, a.Force)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if refreshed {
		ctx.PrintPrimaryOutput("Session refreshed")
		return 0
	} else {
		ctx.PrintPrimaryOutput("Session still valid")
		return 0
	}
}
