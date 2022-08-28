package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"strconv"
	"time"
)

type CLIArgumentsListRecords struct {
	Collection  string
	Raw         bool
	Decoded     bool
	IDOnly      bool
	PrettyPrint bool
	Sort        *string
	Limit       *int
	Offset      *int
	After       *time.Time
}

func NewCLIArgumentsListRecords() *CLIArgumentsListRecords {
	return &CLIArgumentsListRecords{
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

func (a *CLIArgumentsListRecords) Mode() cli.Mode {
	return cli.ModeListRecords
}

func (a *CLIArgumentsListRecords) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient list <collection>", "Get a all records in a collection (use --format to define the format)"},
		{"          (--raw | --decoded | --ids)", "Return raw data, decoded payload, or only IDs"},
		{"          [--after <rfc3339>]", "Return only fields updated after this date"},
		{"          [--sort <sort>]", "Sort the result by (newest|index|oldest)"},
		{"          [--limit <n>]", "Return max <n> elements"},
		{"          [--offset <o>]", "Skip the first <n> elements"},
		{"          [--pretty-print]", "Pretty-Print json in decoded data / payload (if possible)"},
	}
}

func (a *CLIArgumentsListRecords) FullHelp() []string {
	return []string{
		"$> ffsclient list <collection> (--raw | --decoded | --ids) [--after <rfc3339>] [--sort <newest|index|oldest>] [--pretty-print]",
		"",
		"List all records in a collection",
		"",
		"Either --raw or --decoded or --ids must be specified",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest )",
		"The global --format option is used to control the output format",
		"If --pretty-print is specified we try to pretty-print the payload/data, only works if it is in JSON format",
	}
}

func (a *CLIArgumentsListRecords) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) < 1 {
		return errorx.InternalError.New("Not enough arguments for <list> (must be exactly 1)")
	}
	if len(positionalArgs) > 1 {
		return errorx.InternalError.New("Too many arguments for <list> (must be exactly 1)")
	}

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
				return errorx.InternalError.New("Failed to decode time argument '" + arg.Key + "' (expected format: RFC3339)")
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
				return errorx.InternalError.New("Invalid parameter to sort: '" + *arg.Value + "'")
			}
			continue
		}
		if arg.Key == "limit" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Limit = langext.Ptr(int(v))
				continue
			}
			return errorx.InternalError.New("Failed to parse number argument '--limit': '" + *arg.Value + "'")
		}
		if arg.Key == "offset" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Offset = langext.Ptr(int(v))
				continue
			}
			return errorx.InternalError.New("Failed to parse number argument '--offset': '" + *arg.Value + "'")
		}
		if arg.Key == "pretty-print" && arg.Value == nil {
			a.PrettyPrint = true
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsListRecords) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[List_Records]")
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

func (a *CLIArgumentsListRecords) printIDOnly(ctx *cli.FFSContext, records []models.Record) int {
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
			Collections []xmlentry `xml:"Record"`
			XMLName     struct{}   `xml:"Records"`
		}
		data := xml{Collections: make([]xmlentry, 0)}
		for _, v := range records {
			data.Collections = append(data.Collections, xmlentry{ID: v.ID})
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

func (a *CLIArgumentsListRecords) printRaw(ctx *cli.FFSContext, records []models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		for _, v := range records {
			ctx.PrintPrimaryOutput(v.ID)
			ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, v.Payload))
			ctx.PrintPrimaryOutput("")
		}
		return 0

	case cli.OutputFormatJson:
		j := langext.A{}
		for _, v := range records {
			j = append(j, langext.H{
				"id":            v.ID,
				"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				"modified_unix": v.Modified.Unix(),
				"payload":       v.Payload,
			})
		}
		ctx.PrintPrimaryOutputJSON(j)
		return 0

	case cli.OutputFormatXML:
		type xmlentry struct {
			ID           string `xml:"ID,attr"`
			Modified     string `xml:"Modified,attr"`
			ModifiedUnix int64  `xml:"ModifiedUnix,attr"`
			Payload      string `xml:",innerxml"`
		}
		type xml struct {
			Collections []xmlentry `xml:"Record"`
			XMLName     struct{}   `xml:"Records"`
		}
		data := xml{Collections: make([]xmlentry, 0)}
		for _, v := range records {
			data.Collections = append(data.Collections, xmlentry{
				ID:           v.ID,
				Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				ModifiedUnix: v.Modified.Unix(),
				Payload:      a.prettyPrint(ctx, v.Payload),
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

func (a *CLIArgumentsListRecords) printDecoded(ctx *cli.FFSContext, records []models.Record) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		for _, v := range records {
			ctx.PrintPrimaryOutput(v.ID)
			ctx.PrintPrimaryOutput(v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano))
			ctx.PrintPrimaryOutput(a.prettyPrint(ctx, string(v.DecodedData)))
			ctx.PrintPrimaryOutput("")
		}
		return 0

	case cli.OutputFormatJson:
		j := langext.A{}
		for _, v := range records {
			j = append(j, langext.H{
				"id":            v.ID,
				"modified":      v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				"modified_unix": v.Modified.Unix(),
				"data":          string(v.DecodedData),
			})
		}
		ctx.PrintPrimaryOutputJSON(j)
		return 0

	case cli.OutputFormatXML:
		type xmlentry struct {
			ID           string `xml:"ID,attr"`
			Modified     string `xml:"Modified,attr"`
			ModifiedUnix int64  `xml:"ModifiedUnix,attr"`
			Data         string `xml:",innerxml"`
		}
		type xml struct {
			Collections []xmlentry `xml:"Record"`
			XMLName     struct{}   `xml:"Records"`
		}
		data := xml{Collections: make([]xmlentry, 0)}
		for _, v := range records {
			data.Collections = append(data.Collections, xmlentry{
				ID:           v.ID,
				Modified:     v.Modified.In(ctx.Opt.TimeZone).Format(time.RFC3339Nano),
				ModifiedUnix: v.Modified.Unix(),
				Data:         a.prettyPrint(ctx, string(v.DecodedData)),
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

func (a *CLIArgumentsListRecords) prettyPrint(ctx *cli.FFSContext, v string) string {
	if a.PrettyPrint {
		return langext.TryPrettyPrintJson(v)
	} else {
		return v
	}
}
