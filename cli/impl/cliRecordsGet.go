package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"time"
)

type CLIArgumentsRecordsGet struct {
	Collection  string
	RecordID    string
	Raw         bool
	Decoded     bool
	PrettyPrint bool
}

func NewCLIArgumentsRecordsGet() *CLIArgumentsRecordsGet {
	return &CLIArgumentsRecordsGet{
		Raw:         false,
		Decoded:     false,
		PrettyPrint: false,
	}
}

func (a *CLIArgumentsRecordsGet) Mode() cli.Mode {
	return cli.ModeRecordsGet
}

func (a *CLIArgumentsRecordsGet) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsGet) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient get <collection> <record-id>", "Get a single record"},
		{"          (--raw | --decoded)", "Return raw data or decoded payload"},
		{"          [--pretty-print]", "Pretty-Print json in decoded data / payload (if possible)"},
	}
}

func (a *CLIArgumentsRecordsGet) FullHelp() []string {
	return []string{
		"$> ffsclient get <collection> <record-id> (--raw | --decoded) [--pretty-print]",
		"",
		"Get data of a single record",
		"",
		"Either --raw or --decoded must be specified",
		"If --pretty-print is specified we try to pretty-print the payload/data, only works if it is in JSON format",
	}
}

func (a *CLIArgumentsRecordsGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]
	a.RecordID = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "raw" && arg.Value == nil {
			a.Raw = true
			continue
		}
		if arg.Key == "decoded" && arg.Value == nil {
			a.Decoded = true
			continue
		}
		if arg.Key == "pretty-print" && arg.Value == nil {
			a.PrettyPrint = true
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsGet) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Get-Record]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
	ctx.PrintVerboseKV("RecordID", a.RecordID)
	ctx.PrintVerboseKV("RawData", a.Raw)
	ctx.PrintVerboseKV("DecodedData", a.Decoded)

	if !a.Raw && !a.Decoded {
		ctx.PrintFatalMessage("must specify either --raw or --decoded")
		return consts.ExitcodeError
	}
	if a.Raw && a.Decoded {
		ctx.PrintFatalMessage("must specify only one of --raw or --decoded")
		return consts.ExitcodeError
	}

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

	record, err := client.GetRecord(ctx, session, a.Collection, a.RecordID, a.Decoded)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if a.Raw {
		return a.printRaw(ctx, record)
	} else if a.Decoded {
		return a.printDecoded(ctx, record)
	} else {
		ctx.PrintFatalMessage("must specify only one of --raw or --decoded")
		return consts.ExitcodeError
	}
}

func (a *CLIArgumentsRecordsGet) printRaw(ctx *cli.FFSContext, v models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(v.ID)
		ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
		ctx.PrintPrimaryOutput(a.prettyPrint(ctx, v.Payload))
		ctx.PrintPrimaryOutput("")
		return 0

	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(langext.H{
			"id":            v.ID,
			"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
			"modified_unix": v.Modified.Unix(),
			"payload":       v.Payload,
		})
		return 0

	case cli.OutputFormatXML:
		type xml struct {
			ID           string   `xml:"ID,attr"`
			Modified     string   `xml:"Modified,attr"`
			ModifiedUnix int64    `xml:"ModifiedUnix,attr"`
			Payload      string   `xml:",innerxml"`
			XMLName      struct{} `xml:"Record"`
		}
		ctx.PrintPrimaryOutputXML(xml{
			ID:           v.ID,
			Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
			ModifiedUnix: v.Modified.Unix(),
			Payload:      a.prettyPrint(ctx, v.Payload),
		})
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat

	}
}

func (a *CLIArgumentsRecordsGet) printDecoded(ctx *cli.FFSContext, v models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(v.ID)
		ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
		ctx.PrintPrimaryOutput(a.prettyPrint(ctx, string(v.DecodedData)))
		ctx.PrintPrimaryOutput("")
		return 0

	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(langext.H{
			"id":            v.ID,
			"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
			"modified_unix": v.Modified.Unix(),
			"data":          string(v.DecodedData),
		})
		return 0

	case cli.OutputFormatXML:
		type xml struct {
			ID           string   `xml:"ID,attr"`
			Modified     string   `xml:"Modified,attr"`
			ModifiedUnix int64    `xml:"ModifiedUnix,attr"`
			Data         string   `xml:",innerxml"`
			XMLName      struct{} `xml:"Record"`
		}
		ctx.PrintPrimaryOutputXML(xml{
			ID:           v.ID,
			Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
			ModifiedUnix: v.Modified.Unix(),
			Data:         a.prettyPrint(ctx, string(v.DecodedData)),
		})
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat

	}
}

func (a *CLIArgumentsRecordsGet) prettyPrint(ctx *cli.FFSContext, v string) string {
	if a.PrettyPrint {
		return langext.TryPrettyPrintJson(v)
	} else {
		return v
	}
}
