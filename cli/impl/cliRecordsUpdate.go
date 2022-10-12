package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"github.com/joomcode/errorx"
)

type CLIArgumentsRecordsUpdate struct {
	Collection                string
	RecordID                  string
	RawPayload                *string
	DecryptedPayload          *string
	RawPayloadFromStdIn       bool
	DecryptedPayloadFromStdIn bool
	CreateIfNotExistant       bool
	
	CLIArgumentsRecordsUtil
}

func NewCLIArgumentsRecordsUpdate() *CLIArgumentsRecordsUpdate {
	return &CLIArgumentsRecordsUpdate{
		Collection:                "",
		RecordID:                  "",
		RawPayload:                nil,
		DecryptedPayload:          nil,
		RawPayloadFromStdIn:       false,
		DecryptedPayloadFromStdIn: false,
		CreateIfNotExistant:       false,
	}
}

func (a *CLIArgumentsRecordsUpdate) Mode() cli.Mode {
	return cli.ModeRecordsUpdate
}

func (a *CLIArgumentsRecordsUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsUpdate) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsRecordsUpdate) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient update <collection> <record-id>", "Update an existing record"},
		{"          (--raw <r> | --data <d> | --raw-stdin | --data-stdin)", "The new data"},
		{"          [--create]", "Create a new record if the specified record-id does not exist"},
	}
}

func (a *CLIArgumentsRecordsUpdate) FullHelp() []string {
	return []string{
		"$> ffsclient update <collection> <record-id> (--raw <raw> | --data <data> | --raw-stdin | --data-stdin) [--create]",
		"",
		"Update an existing record",
		"",
		"The payload can either be specified:",
		" - directly with --raw <...>",
		" - as unencrypted data with --data <...> (which is then encrypted before written to the server)",
		" - read as raw data from stdin with --raw-stdin",
		" - read as unencrypted data from stdin with --data-stdin (which is then encrypted before written to the server)",
		"If --create is not supplied we test first if the record exists on the server, otherwise we directly call PUT on the server.",
		"This means that a call with --create is faster than without, because the extra check step is no longer needed.",
		"Also this command is suspectible to race-conditions, because two clients can create a record simuatneously and the later will override the first.",
	}
}

func (a *CLIArgumentsRecordsUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "raw" && arg.Value != nil {
			a.RawPayload = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "raw-stdin" && arg.Value == nil {
			a.RawPayloadFromStdIn = true
			continue
		}
		if arg.Key == "data" && arg.Value != nil {
			a.DecryptedPayload = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "data-stdin" && arg.Value == nil {
			a.DecryptedPayloadFromStdIn = true
			continue
		}
		if arg.Key == "create" && arg.Value == nil {
			a.CreateIfNotExistant = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsUpdate) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Update Record]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
	ctx.PrintVerboseKV("RecordID", a.RecordID)
	ctx.PrintVerboseKV("Data<Raw>", a.RawPayload != nil)
	ctx.PrintVerboseKV("Data<Data>", a.DecryptedPayload != nil)
	ctx.PrintVerboseKV("Data<Raw-stdin>", a.RawPayloadFromStdIn)
	ctx.PrintVerboseKV("Data<Data-stdin>", a.DecryptedPayloadFromStdIn)

	if langext.BoolCount(a.RawPayload != nil, a.DecryptedPayload != nil, a.RawPayloadFromStdIn, a.DecryptedPayloadFromStdIn) == 0 {
		return fferr.NewDirectOutput(consts.ExitcodeError, "Must specify one of --raw, --data, --raw-stdin or --data-stdin")
	}
	if langext.BoolCount(a.RawPayload != nil, a.DecryptedPayload != nil, a.RawPayloadFromStdIn, a.DecryptedPayloadFromStdIn) > 1 {
		return fferr.NewDirectOutput(consts.ExitcodeError, "Must specify one of --raw, --data, --raw-stdin or --data-stdin")
	}

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	var payload string

	if a.RawPayload != nil {
		payload = *a.RawPayload
	} else if a.RawPayloadFromStdIn {
		payload, err = ctx.ReadStdIn()
		if err != nil {
			return err
		}
	} else if a.DecryptedPayload != nil {
		payload, err = client.EncryptPayload(ctx, session, a.Collection, *a.DecryptedPayload)
		if err != nil {
			return err
		}
	} else if a.DecryptedPayloadFromStdIn {
		stdin, err := ctx.ReadStdIn()
		if err != nil {
			return err
		}
		payload, err = client.EncryptPayload(ctx, session, a.Collection, stdin)
		if err != nil {
			return err
		}
	}

	// ========================================================================

	update := models.RecordUpdate{
		ID:      a.RecordID,
		Payload: langext.Ptr(payload),
	}

	err = client.PutRecord(ctx, session, a.Collection, update, false, !a.CreateIfNotExistant)
	if err != nil && errorx.IsOfType(err, fferr.Request404) {
		return fferr.NewDirectOutput(consts.ExitcodeRecordNotFound, "Record not found")
	}
	if err != nil {
		return err
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput(a.RecordID)
	return nil
}
