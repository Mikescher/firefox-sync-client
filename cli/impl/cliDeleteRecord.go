package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsDeleteSingle struct {
	Collection string
	RecordID   string
}

func NewCLIArgumentsDeleteSingle() *CLIArgumentsDeleteSingle {
	return &CLIArgumentsDeleteSingle{}
}

func (a *CLIArgumentsDeleteSingle) Mode() cli.Mode {
	return cli.ModeDeleteRecord
}

func (a *CLIArgumentsDeleteSingle) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient delete <collection> <record-id>", "Delete the specified record"},
	}
}

func (a *CLIArgumentsDeleteSingle) FullHelp() []string {
	return []string{
		"$> ffsclient delete <collection> <record-id>",
		"",
		"Delete the specific record from the server",
	}
}

func (a *CLIArgumentsDeleteSingle) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) < 2 {
		return errorx.InternalError.New("Not enough arguments for <delete> (must be exactly 2)")
	}
	if len(positionalArgs) > 2 {
		return errorx.InternalError.New("Too many arguments for <delete> (must be exactly 2)")
	}

	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsDeleteSingle) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete-Record]")
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