package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"github.com/joomcode/errorx"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"time"
)

type CLIArgumentsRecordsGet struct {
	Collection  string
	RecordID    string
	Raw         bool
	Decoded     bool
	PrettyPrint bool
	DataOnly    bool

	CLIArgumentsRecordsUtil
}

func NewCLIArgumentsRecordsGet() *CLIArgumentsRecordsGet {
	return &CLIArgumentsRecordsGet{
		Raw:         false,
		Decoded:     false,
		PrettyPrint: false,
		DataOnly:    false,
	}
}

func (a *CLIArgumentsRecordsGet) Mode() cli.Mode {
	return cli.ModeRecordsGet
}

func (a *CLIArgumentsRecordsGet) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsRecordsGet) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
}

func (a *CLIArgumentsRecordsGet) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient get <collection> <record-id>", "Get a single record"},
		{"          (--raw | --decoded)", "Return raw data or decoded payload"},
		{"          [--pretty-print | --pp]", "Pretty-Print json in decoded data / payload (if possible)"},
		{"          [--data-only]", "Only return the payload"},
	}
}

func (a *CLIArgumentsRecordsGet) FullHelp() []string {
	return []string{
		"$> ffsclient get <collection> <record-id> (--raw | --decoded) [--pretty-print | --pp] [--data-only]",
		"",
		"Get data of a single record",
		"",
		"Either --raw or --decoded must be specified",
		"If --pretty-print is specified we try to pretty-print the payload/data, only works if it is in JSON format",
		"By default we return the 'envelope' around the payload with meta information like id and modification-data.",
		"Use --data-only to only return the (decoded or raw) payload",
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
		if (arg.Key == "pretty-print" || arg.Key == "pp") && arg.Value == nil {
			a.PrettyPrint = true
			continue
		}
		if (arg.Key == "data-only") && arg.Value == nil {
			a.DataOnly = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsGet) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Get Record]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
	ctx.PrintVerboseKV("RecordID", a.RecordID)
	ctx.PrintVerboseKV("RawData", a.Raw)
	ctx.PrintVerboseKV("DecodedData", a.Decoded)

	if !a.Raw && !a.Decoded {
		return fferr.NewDirectOutput(consts.ExitcodeError, "must specify either --raw or --decoded")
	}
	if a.Raw && a.Decoded {
		return fferr.NewDirectOutput(consts.ExitcodeError, "must specify only one of --raw or --decoded")
	}

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	record, err := client.GetRecord(ctx, session, a.Collection, a.RecordID, a.Decoded)
	if err != nil && errorx.IsOfType(err, fferr.Request404) {
		return fferr.NewDirectOutput(consts.ExitcodeRecordNotFound, "Record not found")
	}
	if err != nil {
		return err
	}

	// ========================================================================

	if a.Raw {
		return a.printRaw(ctx, record)
	} else if a.Decoded {
		return a.printDecoded(ctx, record)
	} else {
		return fferr.NewDirectOutput(consts.ExitcodeError, "must specify only one of --raw or --decoded")
	}
}

func (a *CLIArgumentsRecordsGet) printRaw(ctx *cli.FFSContext, v models.Record) error {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		if a.DataOnly {
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, a.PrettyPrint, v.Payload, false))
			ctx.PrintPrimaryOutput("")
		} else {
			ctx.PrintPrimaryOutput(v.ID)
			ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, a.PrettyPrint, v.Payload, false))
			ctx.PrintPrimaryOutput("")
		}
		return nil

	case cli.OutputFormatJson:
		if a.DataOnly {
			ctx.PrintPrimaryOutputJSON(langext.H{"payload": v.Payload})
		} else {
			ctx.PrintPrimaryOutputJSON(langext.H{
				"id":            v.ID,
				"ttl":           v.TTL,
				"sortIndex":     v.SortIndex,
				"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				"modified_unix": v.ModifiedUnix,
				"payload":       v.Payload,
			})
		}
		return nil

	case cli.OutputFormatXML:
		if a.DataOnly {
			type xml struct {
				Payload string   `xml:",chardata"`
				XMLName struct{} `xml:"Data"`
			}
			ctx.PrintPrimaryOutputXML(xml{Payload: a.prettyPrint(ctx, a.PrettyPrint, v.Payload, true)})
		} else {
			type xml struct {
				ID           string   `xml:"ID,attr"`
				TTL          string   `xml:"TTL,attr,omitempty"`
				SortIndex    int64    `xml:"SortIndex,attr"`
				Modified     string   `xml:"Modified,attr"`
				ModifiedUnix float64  `xml:"ModifiedUnix,attr"`
				Payload      string   `xml:",chardata"`
				XMLName      struct{} `xml:"Record"`
			}
			ctx.PrintPrimaryOutputXML(xml{
				ID:           v.ID,
				TTL:          langext.NumToStringOpt(v.TTL, ""),
				SortIndex:    v.SortIndex,
				Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				ModifiedUnix: v.ModifiedUnix,
				Payload:      a.prettyPrint(ctx, a.PrettyPrint, v.Payload, true),
			})
		}
		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())

	}
}

func (a *CLIArgumentsRecordsGet) printDecoded(ctx *cli.FFSContext, v models.Record) error {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		if a.DataOnly {
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, a.PrettyPrint, string(v.DecodedData), false))
			ctx.PrintPrimaryOutput("")
		} else {
			ctx.PrintPrimaryOutput(v.ID)
			ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, a.PrettyPrint, string(v.DecodedData), false))
			ctx.PrintPrimaryOutput("")
		}
		return nil

	case cli.OutputFormatJson:
		if a.DataOnly {
			if a.PrettyPrint {
				ctx.PrintPrimaryOutputJSON(langext.H{"data": a.tryParseJson(ctx, v.DecodedData)})
			} else {
				ctx.PrintPrimaryOutputJSON(langext.H{"data": string(v.DecodedData)})
			}
		} else {
			if a.PrettyPrint {
				ctx.PrintPrimaryOutputJSON(langext.H{
					"id":            v.ID,
					"ttl":           v.TTL,
					"sortIndex":     v.SortIndex,
					"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
					"modified_unix": v.ModifiedUnix,
					"data":          a.tryParseJson(ctx, v.DecodedData),
				})
			} else {
				ctx.PrintPrimaryOutputJSON(langext.H{
					"id":            v.ID,
					"ttl":           v.TTL,
					"sortIndex":     v.SortIndex,
					"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
					"modified_unix": v.ModifiedUnix,
					"data":          string(v.DecodedData),
				})
			}
		}
		return nil

	case cli.OutputFormatXML:
		if a.DataOnly {
			type xml struct {
				Data    string   `xml:",chardata"`
				XMLName struct{} `xml:"Record"`
			}
			ctx.PrintPrimaryOutputXML(xml{
				Data: a.prettyPrint(ctx, a.PrettyPrint, string(v.DecodedData), true),
			})
		} else {
			type xml struct {
				ID           string   `xml:"ID,attr"`
				TTL          string   `xml:"TTL,attr,omitempty"`
				SortIndex    int64    `xml:"SortIndex,attr"`
				Modified     string   `xml:"Modified,attr"`
				ModifiedUnix float64  `xml:"ModifiedUnix,attr"`
				Data         string   `xml:",chardata"`
				XMLName      struct{} `xml:"Record"`
			}
			ctx.PrintPrimaryOutputXML(xml{
				ID:           v.ID,
				TTL:          langext.NumToStringOpt(v.TTL, ""),
				SortIndex:    v.SortIndex,
				Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				ModifiedUnix: v.ModifiedUnix,
				Data:         a.prettyPrint(ctx, a.PrettyPrint, string(v.DecodedData), true),
			})
		}
		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())

	}
}
