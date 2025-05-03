package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsDeleteAll struct {
	Force bool
	CLIArgumentsBaseUtil
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

	client, session, err := a.InitClient(ctx)
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
