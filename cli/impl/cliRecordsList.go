package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"strconv"
	"time"
)

type CLIArgumentsRecordsList struct {
	Collection  string
	Raw         bool
	Decoded     bool
	IDOnly      bool
	PrettyPrint bool
	Sort        *string
	Limit       *int
	Offset      *int
	After       *time.Time

	CLIArgumentsRecordsUtil
}

func NewCLIArgumentsRecordsList() *CLIArgumentsRecordsList {
	return &CLIArgumentsRecordsList{
		Raw:         false,
		Decoded:     false,
		IDOnly:      false,
		PrettyPrint: false,
		Sort:        nil,
		Limit:       nil,
		Offset:      nil,
		After:       nil,
	}
}

func (a *CLIArgumentsRecordsList) Mode() cli.Mode {
	return cli.ModeRecordsList
}

func (a *CLIArgumentsRecordsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsRecordsList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient list <collection>", "Get a all records in a collection (use --format to define the format)"},
		{"          (--raw | --decoded | --ids)", "Return raw data, decoded payload, or only IDs"},
		{"          [--after <rfc3339>]", "Return only fields updated after this date"},
		{"          [--sort <sort>]", "Sort the result by (newest|index|oldest)"},
		{"          [--limit <n>]", "Return max <n> elements"},
		{"          [--offset <o>]", "Skip the first <n> elements"},
		{"          [--pretty-print | --pp]", "Pretty-Print json in decoded data / payload (if possible)"},
	}
}

func (a *CLIArgumentsRecordsList) FullHelp() []string {
	return []string{
		"$> ffsclient list <collection> (--raw | --decoded | --ids) [--after <rfc3339>] [--sort <newest|index|oldest>] [--pretty-print | --pp]",
		"",
		"List all records in a collection",
		"",
		"Either --raw or --decoded or --ids must be specified",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest )",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
		"The global --format option is used to control the output format",
		"If --pretty-print is specified we try to pretty-print the payload/data, only works if it is in JSON format.",
		"If the output-format is json and we specify --pretty-print the output json also contains the raw data-json (instead of an string-enncoded version)",
	}
}

func (a *CLIArgumentsRecordsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Collection = positionalArgs[0]

	for _, arg := range optionArgs {
		if arg.Key == "raw" && arg.Value == nil {
			a.Raw = true
			continue
		}
		if arg.Key == "decoded" && arg.Value == nil {
			a.Decoded = true
			continue
		}
		if arg.Key == "ids" && arg.Value == nil {
			a.IDOnly = true
			continue
		}
		if arg.Key == "after" && arg.Value != nil {
			if t, err := time.Parse(time.RFC3339Nano, *arg.Value); err == nil {
				a.After = langext.Ptr(t)
			} else if t, err := time.Parse(time.RFC3339, *arg.Value); err == nil {
				a.After = langext.Ptr(t)
			} else {
				return fferr.DirectOutput.New("Failed to decode time argument '" + arg.Key + "' (expected format: RFC3339)")
			}
			continue
		}
		if arg.Key == "sort" && arg.Value != nil {
			if *arg.Value == "newest" {
				a.Sort = langext.Ptr("newest")
			} else if *arg.Value == "index" {
				a.Sort = langext.Ptr("index")
			} else if *arg.Value == "oldest" {
				a.Sort = langext.Ptr("oldest")
			} else {
				return fferr.DirectOutput.New("Invalid parameter for sort: '" + *arg.Value + "'")
			}
			continue
		}
		if arg.Key == "limit" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Limit = langext.Ptr(int(v))
				continue
			}
			return fferr.DirectOutput.New("Failed to parse number argument '--limit': '" + *arg.Value + "'")
		}
		if arg.Key == "offset" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Offset = langext.Ptr(int(v))
				continue
			}
			return fferr.DirectOutput.New("Failed to parse number argument '--offset': '" + *arg.Value + "'")
		}
		if (arg.Key == "pretty-print" || arg.Key == "pp") && arg.Value == nil {
			a.PrettyPrint = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsRecordsList) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[List Records]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Collection", a.Collection)
	ctx.PrintVerboseKV("RawData", a.Raw)
	ctx.PrintVerboseKV("DecodedData", a.Decoded)
	ctx.PrintVerboseKV("IDOnly", a.IDOnly)

	if !a.Raw && !a.Decoded && !a.IDOnly {
		ctx.PrintFatalMessage("must specify either --raw or --decoded or --ids")
		return consts.ExitcodeError
	}
	if (a.Raw && a.Decoded) || (a.Decoded && a.IDOnly) || (a.IDOnly && a.Raw) {
		ctx.PrintFatalMessage("must specify only one of --raw or --decoded or --ids")
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

	records, err := client.ListRecords(ctx, session, a.Collection, a.After, a.Sort, a.IDOnly, a.Decoded, a.Limit, a.Offset)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	if a.IDOnly {
		return a.printIDOnly(ctx, records)
	} else if a.Raw {
		return a.printRaw(ctx, records)
	} else if a.Decoded {
		return a.printDecoded(ctx, records)
	} else {
		ctx.PrintFatalMessage("must specify only one of --raw or --decoded or --ids")
		return consts.ExitcodeError
	}
}

func (a *CLIArgumentsRecordsList) printIDOnly(ctx *cli.FFSContext, records []models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		for _, v := range records {
			ctx.PrintPrimaryOutput(v.ID)
		}
		return 0

	case cli.OutputFormatJson:
		arr := make([]string, 0, len(records))
		for _, v := range records {
			arr = append(arr, v.ID)
		}
		ctx.PrintPrimaryOutputJSON(arr)
		return 0

	case cli.OutputFormatXML:
		type xmlentry struct {
			ID string `xml:",innerxml"`
		}
		type xml struct {
			Records []xmlentry `xml:"Record"`
			XMLName struct{}   `xml:"Records"`
		}
		data := xml{Records: make([]xmlentry, 0)}
		for _, v := range records {
			data.Records = append(data.Records, xmlentry{ID: v.ID})
		}
		ctx.PrintPrimaryOutputXML(data)
		return 0

	case cli.OutputFormatTable:
		for _, v := range records {
			ctx.PrintPrimaryOutput(v.ID)
		}
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat

	}
}

func (a *CLIArgumentsRecordsList) printRaw(ctx *cli.FFSContext, records []models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		for _, v := range records {
			ctx.PrintPrimaryOutput(v.ID)
			ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, a.PrettyPrint, v.Payload, false))
			ctx.PrintPrimaryOutput("")
		}
		return 0

	case cli.OutputFormatJson:
		j := langext.A{}
		for _, v := range records {
			j = append(j, langext.H{
				"id":            v.ID,
				"ttl":           v.TTL,
				"sortIndex":     v.SortIndex,
				"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				"modified_unix": v.ModifiedUnix,
				"payload":       v.Payload,
			})
		}
		ctx.PrintPrimaryOutputJSON(j)
		return 0

	case cli.OutputFormatXML:
		type xmlentry struct {
			ID           string  `xml:"ID,attr"`
			TTL          string  `xml:"TTL,attr,omitempty"`
			SortIndex    int64   `xml:"SortIndex,attr"`
			Modified     string  `xml:"Modified,attr"`
			ModifiedUnix float64 `xml:"ModifiedUnix,attr"`
			Payload      string  `xml:",innerxml"`
		}
		type xml struct {
			Records []xmlentry `xml:"Record"`
			XMLName struct{}   `xml:"Records"`
		}
		data := xml{Records: make([]xmlentry, 0)}
		for _, v := range records {
			data.Records = append(data.Records, xmlentry{
				ID:           v.ID,
				TTL:          langext.NumToStringOpt(v.TTL, ""),
				SortIndex:    v.SortIndex,
				Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				ModifiedUnix: v.ModifiedUnix,
				Payload:      a.prettyPrint(ctx, a.PrettyPrint, v.Payload, true),
			})
		}
		ctx.PrintPrimaryOutputXML(data)
		return 0

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(records))
		table = append(table, []string{"ID", "LAST MODIFIED", "PAYLOAD"})
		for _, v := range records {
			table = append(table, []string{
				v.ID,
				v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				v.Payload,
			})
		}

		ctx.PrintPrimaryOutputTable(table, true)
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat

	}
}

func (a *CLIArgumentsRecordsList) printDecoded(ctx *cli.FFSContext, records []models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		for _, v := range records {
			ctx.PrintPrimaryOutput(v.ID)
			ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, a.PrettyPrint, string(v.DecodedData), false))
			ctx.PrintPrimaryOutput("")
		}
		return 0

	case cli.OutputFormatJson:
		j := langext.A{}
		for _, v := range records {
			if a.PrettyPrint {
				j = append(j, langext.H{
					"id":            v.ID,
					"ttl":           v.TTL,
					"sortIndex":     v.SortIndex,
					"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
					"modified_unix": v.ModifiedUnix,
					"data":          a.tryParseJson(ctx, v.DecodedData),
				})
			} else {
				j = append(j, langext.H{
					"id":            v.ID,
					"ttl":           v.TTL,
					"sortIndex":     v.SortIndex,
					"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
					"modified_unix": v.ModifiedUnix,
					"data":          string(v.DecodedData),
				})
			}
		}
		ctx.PrintPrimaryOutputJSON(j)
		return 0

	case cli.OutputFormatXML:
		type xmlentry struct {
			ID           string  `xml:"ID,attr"`
			TTL          string  `xml:"TTL,attr,omitempty"`
			SortIndex    int64   `xml:"SortIndex,attr"`
			Modified     string  `xml:"Modified,attr"`
			ModifiedUnix float64 `xml:"ModifiedUnix,attr"`
			Data         string  `xml:",innerxml"`
		}
		type xml struct {
			Records []xmlentry `xml:"Record"`
			XMLName struct{}   `xml:"Records"`
		}
		data := xml{Records: make([]xmlentry, 0)}
		for _, v := range records {
			data.Records = append(data.Records, xmlentry{
				ID:           v.ID,
				TTL:          langext.NumToStringOpt(v.TTL, ""),
				SortIndex:    v.SortIndex,
				Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				ModifiedUnix: v.ModifiedUnix,
				Data:         a.prettyPrint(ctx, a.PrettyPrint, string(v.DecodedData), true),
			})
		}
		ctx.PrintPrimaryOutputXML(data)
		return 0

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(records))
		table = append(table, []string{"ID", "LAST MODIFIED", "DATA"})
		for _, v := range records {
			table = append(table, []string{
				v.ID,
				v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				string(v.DecodedData),
			})
		}

		ctx.PrintPrimaryOutputTable(table, true)
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat

	}
}
