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
}

func NewCLIArgumentsRecordsDelete() *CLIArgumentsRecordsDelete {
	return &CLIArgumentsRecordsDelete{}
}

func (a *CLIArgumentsRecordsDelete) Mode() cli.Mode {
	return cli.ModeRecordsDelete
}

func (a *CLIArgumentsRecordsDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient delete <collection> <record-id>", "Delete the specified record"},
	}
}

func (a *CLIArgumentsRecordsDelete) FullHelp() []string {
	return []string{
		"$> ffsclient delete <collection> <record-id>",
		"",
		"Delete the specific record from the server",
	}
}

func (a *CLIArgumentsRecordsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsDelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete Record]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
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

	err = client.DeleteRecord(ctx, session, a.Collection, a.RecordID)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput("Record " + a.RecordID + " deleted")
	return 0
}
