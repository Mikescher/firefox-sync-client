package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
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

func (a *CLIArgumentsHistoryDelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete History]")
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

		err = client.DeleteRecord(ctx, session, consts.CollectionHistory, a.RecordID)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			ctx.PrintErrorMessage("Record not found")
			return consts.ExitcodeRecordNotFound
		}
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, consts.CollectionHistory, a.RecordID)
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
