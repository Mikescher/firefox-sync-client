package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsDeleteCollection struct {
	Collection string
}

func NewCLIArgumentsDeleteCollection() *CLIArgumentsDeleteCollection {
	return &CLIArgumentsDeleteCollection{}
}

func (a *CLIArgumentsDeleteCollection) Mode() cli.Mode {
	return cli.ModeDeleteCollection
}

func (a *CLIArgumentsDeleteCollection) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient delete <collection>", "Delete the all records in a collection"},
	}
}

func (a *CLIArgumentsDeleteCollection) FullHelp() []string {
	return []string{
		"$> ffsclient delete <collection>",
		"",
		"Delete the all records in a collection",
	}
}

func (a *CLIArgumentsDeleteCollection) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) < 1 {
		return errorx.InternalError.New("Not enough arguments for <delete> (must be exactly 2)")
	}
	if len(positionalArgs) > 1 {
		return errorx.InternalError.New("Too many arguments for <delete> (must be exactly 2)")
	}

	a.Collection = positionalArgs[0]

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsDeleteCollection) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete-Collection]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)

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

	err = client.DeleteCollection(ctx, session, a.Collection)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput("Collection " + a.Collection + " deleted")
	return 0
}
