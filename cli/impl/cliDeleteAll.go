package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
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

func (a *CLIArgumentsDeleteAll) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
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
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsDeleteAll) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete Data]")
	ctx.PrintVerbose("")

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return err
	}

	if !langext.FileExists(cfp) {
		return fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return err
	}

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		return err
	}

	// ========================================================================

	if !a.Force {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "The delete-all command needs the --force flag")
	}

	err = client.DeleteAllData(ctx, session)
	if err != nil {
		return err
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput("Data deleted")
	return nil
}
