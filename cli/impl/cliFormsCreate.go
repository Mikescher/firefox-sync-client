package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsCreate struct {
	Name  string
	Value string

	CLIArgumentsFormsUtil
}

func NewCLIArgumentsFormsCreate() *CLIArgumentsFormsCreate {
	return &CLIArgumentsFormsCreate{
		CLIArgumentsFormsUtil: CLIArgumentsFormsUtil{},
	}
}

func (a *CLIArgumentsFormsCreate) Mode() cli.Mode {
	return cli.ModeFormsCreate
}

func (a *CLIArgumentsFormsCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsFormsCreate) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsFormsCreate) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient forms create <name> <value>", "Adds a new HTML-Form autocomplete suggestions"},
	}
}

func (a *CLIArgumentsFormsCreate) FullHelp() []string {
	return []string{
		"$> ffsclient forms create <name> <value>",
		"",
		"Adds a new HTML-Form autocomplete suggestions",
		"",
		"Outputs the ID of the created entry.",
	}
}

func (a *CLIArgumentsFormsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Name = positionalArgs[0]
	a.Value = positionalArgs[1]

	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsCreate) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Create Bookmark<Folder>]")
	ctx.PrintVerbose("")

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

	recordID := a.newFormID()

	ctx.PrintVerboseHeader("[1] Create new record")

	bso := models.FormPayloadSchema{
		ID:    recordID,
		Name:  a.Name,
		Value: a.Value,
	}

	plainPayload, err := json.Marshal(bso)
	if err != nil {
		ctx.PrintFatalError(errorx.Decorate(err, "failed to marshal BSO json"))
		return consts.ExitcodeError
	}

	payloadNewRecord, err := client.EncryptPayload(ctx, session, consts.CollectionForms, string(plainPayload))
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	update := models.RecordUpdate{
		ID:      recordID,
		Payload: langext.Ptr(payloadNewRecord),
	}

	err = client.PutRecord(ctx, session, consts.CollectionForms, update, true, false)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput(recordID)
	return 0
}
