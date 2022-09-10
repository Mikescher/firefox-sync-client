package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
)

type CLIArgumentsRecordsDelete struct {
	Collection string
	RecordID   string
	HardDelete bool
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

func (a *CLIArgumentsRecordsDelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete Record]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
	ctx.PrintVerboseKV("RecordID", a.RecordID)
	ctx.PrintVerboseKV("HardDelete", a.HardDelete)

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

		err = client.DeleteRecord(ctx, session, a.Collection, a.RecordID)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, a.Collection, a.RecordID)
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

	ctx.PrintPrimaryOutput("Record " + a.RecordID + " deleted")
	return 0
}
