package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsDelete struct {
	RecordID   string
	HardDelete bool

	CLIArgumentsFormsUtil
}

func NewCLIArgumentsFormsDelete() *CLIArgumentsFormsDelete {
	return &CLIArgumentsFormsDelete{}
}

func (a *CLIArgumentsFormsDelete) Mode() cli.Mode {
	return cli.ModeFormsDelete
}

func (a *CLIArgumentsFormsDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsFormsDelete) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsFormsDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient forms delete <id> [--hard]", "Delete the specified HTML-Form autocomplete suggestion"},
	}
}

func (a *CLIArgumentsFormsDelete) FullHelp() []string {
	return []string{
		"$> ffsclient forms delete <id> [--hard]",
		"",
		"Delete the specific HTML-Form autocomplete suggestion from the server",
		"If --hard is specified we delete the record, otherwise we only add {deleted:true} to mark it as a tombstone",
	}
}

func (a *CLIArgumentsFormsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.RecordID = positionalArgs[0]

	for _, arg := range optionArgs {
		if arg.Key == "hard" && arg.Value == nil {
			a.HardDelete = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsDelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete Form]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("RecordID", a.RecordID)

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

	if a.HardDelete {

		err = client.DeleteRecord(ctx, session, consts.CollectionForms, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			ctx.PrintErrorMessage("Record not found")
			return consts.ExitcodeRecordNotFound
		}
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, consts.CollectionForms, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			ctx.PrintErrorMessage("Record not found")
			return consts.ExitcodeRecordNotFound
		}
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	if a.HardDelete {
		ctx.PrintPrimaryOutput("Entry " + a.RecordID + " deleted")
	} else {
		ctx.PrintPrimaryOutput("Entry " + a.RecordID + " marked as deleted")
	}
	return 0
}
