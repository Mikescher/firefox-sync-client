package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
)

type CLIArgumentsRecordsCreate struct {
	Collection                string
	RecordID                  string
	RawPayload                *string
	DecryptedPayload          *string
	RawPayloadFromStdIn       bool
	DecryptedPayloadFromStdIn bool

	CLIArgumentsRecordsUtil
}

func NewCLIArgumentsRecordsCreate() *CLIArgumentsRecordsCreate {
	return &CLIArgumentsRecordsCreate{
		Collection:                "",
		RecordID:                  "",
		RawPayload:                nil,
		DecryptedPayload:          nil,
		RawPayloadFromStdIn:       false,
		DecryptedPayloadFromStdIn: false,
	}
}

func (a *CLIArgumentsRecordsCreate) Mode() cli.Mode {
	return cli.ModeRecordsCreate
}

func (a *CLIArgumentsRecordsCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsCreate) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsRecordsCreate) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient create <collection> <record-id>", "Insert a new record"},
		{"          (--raw <r> | --data <d> | --raw-stdin | --data-stdin)", "The new data"},
	}
}

func (a *CLIArgumentsRecordsCreate) FullHelp() []string {
	return []string{
		"$> ffsclient create <collection> <record-id> (--raw <raw> | --data <data> | --raw-stdin | --data-stdin)",
		"",
		"Insert a new record",
		"",
		"The payload can either be specified:",
		" - directly with --raw <...>",
		" - as unencrypted data with --data <...> (which is then encrypted before written to the server)",
		" - read as raw data from stdin with --raw-stdin",
		" - read as unencrypted data from stdin with --data-stdin (which is then encrypted before written to the server)",
		"The Record ID must be a new unique identifier (use for example `uuidgen`)",
		"If you want to upsert a record, use `ffsclient update --create` (see `ffsclient update --help`)",
	}
}

func (a *CLIArgumentsRecordsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsCreate) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Create Record]")
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
		return fferr.NewDirectOutput(consts.ExitcodeError, "Must specify at most one of --raw, --data, --raw-stdin or --data-stdin")
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

	update := models.RecordUpdate{
		ID:      a.RecordID,
		Payload: langext.Ptr(payload),
	}

	err = client.PutRecord(ctx, session, a.Collection, update, true, false)
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
