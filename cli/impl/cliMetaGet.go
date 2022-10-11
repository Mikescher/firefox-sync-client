package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
)

type CLIArgumentsMetaGet struct {
}

func NewCLIArgumentsMetaGet() *CLIArgumentsMetaGet {
	return &CLIArgumentsMetaGet{}
}

func (a *CLIArgumentsMetaGet) Mode() cli.Mode {
	return cli.ModeMetaGet
}

func (a *CLIArgumentsMetaGet) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsMetaGet) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsMetaGet) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient meta", "Get storage metadata"},
	}
}

func (a *CLIArgumentsMetaGet) FullHelp() []string {
	return []string{
		"$> ffsclient meta",
		"",
		"Get storage metadata",
	}
}

func (a *CLIArgumentsMetaGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsMetaGet) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Get Meta]")
	ctx.PrintVerbose("")

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return err
	}

	if !langext.FileExists(cfp) {
		return fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return err
	}

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		return err
	}

	// ========================================================================

	record, err := client.GetRecord(ctx, session, consts.CollectionMeta, consts.RecordMetaGlobal, false)
	if err != nil {
		return err
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput(langext.TryPrettyPrintJson(record.Payload))
	return nil
}
