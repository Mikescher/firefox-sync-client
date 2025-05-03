package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsMetaGet struct {
	CLIArgumentsBaseUtil
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

	client, session, err := a.InitClient(ctx)
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
