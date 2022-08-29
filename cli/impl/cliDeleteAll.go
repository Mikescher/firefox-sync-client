package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsDeleteAll struct {
	Force bool
}

func NewCLIArgumentsDeleteAll() *CLIArgumentsDeleteAll {
	return &CLIArgumentsDeleteAll{}
}

func (a *CLIArgumentsDeleteAll) Mode() cli.Mode {
	return cli.ModeDeleteAll
}

func (a *CLIArgumentsDeleteAll) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsDeleteAll) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient delete-all --force", "Delete all (!) records in the server"},
	}
}

func (a *CLIArgumentsDeleteAll) FullHelp() []string {
	return []string{
		"$> ffsclient delete-all",
		"",
		"Delete the all records on the server",
		"",
		"The --force flag is required",
		"Warning (!): This also deletes the crypto/keys record and can mess with further use of the account",
	}
}

func (a *CLIArgumentsDeleteAll) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "force" && arg.Value == nil {
			a.Force = true
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsDeleteAll) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete-Collection]")
	ctx.PrintVerbose("")

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

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if !a.Force {
		ctx.PrintFatalMessage("The delete-all command needs the --force flag")
		return consts.ExitcodeUnsupportedOutputFormat
	}

	err = client.DeleteAllData(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput("Data deleted")
	return 0
}
