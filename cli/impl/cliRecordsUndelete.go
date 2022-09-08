package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
)

type CLIArgumentsRecordsUndelete struct {
	Collection string
	RecordID   string
}

func NewCLIArgumentsRecordsUndelete() *CLIArgumentsRecordsUndelete {
	return &CLIArgumentsRecordsUndelete{}
}

func (a *CLIArgumentsRecordsUndelete) Mode() cli.Mode {
	return cli.ModeRecordsUndelete
}

func (a *CLIArgumentsRecordsUndelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsUndelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient undelete <collection> <record-id>", "Un-Deletes the specified record"},
	}
}

func (a *CLIArgumentsRecordsUndelete) FullHelp() []string {
	return []string{
		"$> ffsclient undelete <collection> <record-id>",
		"",
		"Undelete the specific record from the server",
		"THis only works if the record was not hard deleted but only flagged as {deleted:true} (eg. with ffsclient delete --soft)",
	}
}

func (a *CLIArgumentsRecordsUndelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsUndelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Undelete Record]")
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
	jsonpayload["deleted"] = false

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

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput("Record " + a.RecordID + " deleted")
	return 0
}
