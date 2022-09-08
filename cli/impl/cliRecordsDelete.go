package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
)

type CLIArgumentsRecordsDelete struct {
	Collection string
	RecordID   string
	SoftDelete bool
}

func NewCLIArgumentsRecordsDelete() *CLIArgumentsRecordsDelete {
	return &CLIArgumentsRecordsDelete{
		SoftDelete: false,
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
		{"ffsclient delete <collection> <record-id> [--soft]", "Delete the specified record"},
	}
}

func (a *CLIArgumentsRecordsDelete) FullHelp() []string {
	return []string{
		"$> ffsclient delete <collection> <record-id>",
		"",
		"Delete the specific record from the server",
		"If --soft is specified we do not really delete the record, but only add {deleted:true} to its payload",
	}
}

func (a *CLIArgumentsRecordsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "soft" && arg.Value == nil {
			a.SoftDelete = true
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
	ctx.PrintVerboseKV("Soft", a.SoftDelete)

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

	if a.SoftDelete {

		record, err := client.GetRecord(ctx, session, a.Collection, a.RecordID, true)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

		var jsonpayload langext.H
		err = json.Unmarshal(record.DecodedData, &jsonpayload)
		if err != nil {
			ctx.PrintFatalError(fferr.DirectOutput.Wrap(err, "failed to unmarshal"))
			return consts.ExitcodeError
		}
		jsonpayload["deleted"] = true

		plainpayload, err := json.Marshal(jsonpayload)
		if err != nil {
			ctx.PrintFatalError(fferr.DirectOutput.Wrap(err, "failed to re-marshal payload"))
			return consts.ExitcodeError
		}

		payload, err := client.EncryptPayload(ctx, session, a.Collection, string(plainpayload))
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

		err = client.PutRecord(ctx, session, a.Collection, a.RecordID, payload, false, false)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

	} else {

		err = client.DeleteRecord(ctx, session, a.Collection, a.RecordID)
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
