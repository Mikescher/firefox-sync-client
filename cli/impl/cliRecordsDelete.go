package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsRecordsDelete struct {
	Collection string
	RecordID   string
	HardDelete bool

	CLIArgumentsRecordsUtil
}

func NewCLIArgumentsRecordsDelete() *CLIArgumentsRecordsDelete {
	return &CLIArgumentsRecordsDelete{
		HardDelete: false,
	}
}

func (a *CLIArgumentsRecordsDelete) Mode() cli.Mode {
	return cli.ModeRecordsDelete
}

func (a *CLIArgumentsRecordsDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsDelete) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsRecordsDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient delete <collection> <record-id> [--hard]", "Delete the specified record"},
	}
}

func (a *CLIArgumentsRecordsDelete) FullHelp() []string {
	return []string{
		"$> ffsclient delete <collection> <record-id> [--hard]",
		"",
		"Delete the specific record from the server",
		"If --hard is specified we delete the record, otherwise we only add {deleted:true} to mark it as a tombstone",
	}
}

func (a *CLIArgumentsRecordsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "hard" && arg.Value == nil {
			a.HardDelete = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsDelete) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete Record]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
	ctx.PrintVerboseKV("RecordID", a.RecordID)
	ctx.PrintVerboseKV("HardDelete", a.HardDelete)

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	if a.HardDelete {

		err = client.DeleteRecord(ctx, session, a.Collection, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.NewDirectOutput(consts.ExitcodeRecordNotFound, "Record not found")
		}
		if err != nil {
			return err
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, a.Collection, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.NewDirectOutput(consts.ExitcodeRecordNotFound, "Record not found")
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
		ctx.PrintPrimaryOutput("Record " + a.RecordID + " deleted")
	} else {
		ctx.PrintPrimaryOutput("Record " + a.RecordID + " marked as deleted")
	}
	return nil
}
