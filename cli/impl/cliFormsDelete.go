package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
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

func (a *CLIArgumentsFormsDelete) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete Form]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("RecordID", a.RecordID)

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	if a.HardDelete {

		err = client.DeleteRecord(ctx, session, consts.CollectionForms, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.WrapDirectOutput(err, consts.ExitcodeRecordNotFound, "Record not found")
		}
		if err != nil {
			return err
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, consts.CollectionForms, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.WrapDirectOutput(err, consts.ExitcodeRecordNotFound, "Record not found")
		}
		if err != nil {
			return err
		}

	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	if a.HardDelete {
		ctx.PrintPrimaryOutput("Entry " + a.RecordID + " deleted")
	} else {
		ctx.PrintPrimaryOutput("Entry " + a.RecordID + " marked as deleted")
	}
	return nil
}
