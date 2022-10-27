package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsHistoryDelete struct {
	RecordID   string
	HardDelete bool

	CLIArgumentsHistoryUtil
}

func NewCLIArgumentsHistoryDelete() *CLIArgumentsHistoryDelete {
	return &CLIArgumentsHistoryDelete{}
}

func (a *CLIArgumentsHistoryDelete) Mode() cli.Mode {
	return cli.ModeHistoryDelete
}

func (a *CLIArgumentsHistoryDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsHistoryDelete) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsHistoryDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient history delete <id> [--hard]", "Delete the specified history entry"},
	}
}

func (a *CLIArgumentsHistoryDelete) FullHelp() []string {
	return []string{
		"$> ffsclient history delete <id> [--hard]",
		"",
		"Delete the specific history entry from the server",
		"If --hard is specified we delete the record, otherwise we only add {deleted:true} to mark it as a tombstone",
	}
}

func (a *CLIArgumentsHistoryDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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

func (a *CLIArgumentsHistoryDelete) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete History]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("RecordID", a.RecordID)

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	if a.HardDelete {

		err = client.DeleteRecord(ctx, session, consts.CollectionHistory, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.WrapDirectOutput(err, consts.ExitcodeRecordNotFound, "Record not found")
		}
		if err != nil {
			return err
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, consts.CollectionHistory, a.RecordID)
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
