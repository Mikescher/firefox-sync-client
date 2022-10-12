package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"fmt"
	"strconv"
	"time"
)

type CLIArgumentsHistoryList struct {
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time
	IncludeDeleted     bool
	OnlyDeleted        bool

	CLIArgumentsHistoryUtil
}

func NewCLIArgumentsHistoryList() *CLIArgumentsHistoryList {
	return &CLIArgumentsHistoryList{
		IgnoreSchemaErrors: false,
		Sort:               nil,
		Limit:              nil,
		Offset:             nil,
		After:              nil,
		IncludeDeleted:     false,
		OnlyDeleted:        false,
	}
}

func (a *CLIArgumentsHistoryList) Mode() cli.Mode {
	return cli.ModeHistoryList
}

func (a *CLIArgumentsHistoryList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsHistoryList) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
}

func (a *CLIArgumentsHistoryList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient history list", "List form history entries"},
		{"          [--ignore-schema-errors]", "Skip records that cannot be decoded into a history schema"},
		{"          [--after <rfc3339>]", "Return only fields after this date"},
		{"          [--sort <sort>]", "Sort the result by (newest|index|oldest)"},
		{"          [--limit <n>]", "Return max <n> elements"},
		{"          [--offset <o>]", "Skip the first <n> elements"},
		{"          [--include-deleted]", "Show deleted entries"},
		{"          [--only-deleted]", "Show only deleted entries"},
	}
}

func (a *CLIArgumentsHistoryList) FullHelp() []string {
	return []string{
		"$> ffsclient history list [--ignore-schema-errors] [--after <rfc3339>] [--sort <sort>] [--limit <n>] [--offset <o>] [--include-deleted] [--only-deleted]",
		"",
		"List HTML-Form autocomplete suggestions",
		"",
		"If --ignore-schema-errors is not supplied the programm returns with exitcode [0] if any record in the history collection has invalid data. Otherwise we simply skip that record.",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned.",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest ).",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
		"By default we skip entries with {deleted:true}, this can be changed with --include-deleted and --only-deleted.",
	}
}

func (a *CLIArgumentsHistoryList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "ignore-schema-errors" && arg.Value == nil {
			a.IgnoreSchemaErrors = true
			continue
		}
		if arg.Key == "include-deleted" && arg.Value == nil {
			a.IncludeDeleted = true
			continue
		}
		if arg.Key == "only-deleted" && arg.Value == nil {
			a.OnlyDeleted = true
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
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsHistoryList) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[List History Entries]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	records, err := client.ListRecords(ctx, session, consts.CollectionHistory, a.After, a.Sort, false, true, a.Limit, a.Offset)
	if err != nil {
		return err
	}

	entries, err := models.UnmarshalHistories(ctx, records, a.IgnoreSchemaErrors)
	if err != nil {
		return err
	}

	// ========================================================================

	return a.printOutput(ctx, entries)
}

func (a *CLIArgumentsHistoryList) printOutput(ctx *cli.FFSContext, entries []models.HistoryRecord) error {
	entries = a.filterDeleted(ctx, entries, a.IncludeDeleted, a.OnlyDeleted)

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(entries))
		table = append(table, []string{"ID", "DELETED", "URI", "TITLE", "VISITS", "LAST VISIT", "FIRST VISIT"})
		for _, v := range entries {
			table = append(table, []string{
				v.ID,
				langext.FormatBool(v.Deleted, "true", "false"),
				v.URI,
				v.Title,
				fmt.Sprintf("%d", len(v.Visits)),
				v.LastVisitStr(ctx),
				v.FirstVisitStr(ctx),
			})
		}

		if a.IncludeDeleted && !a.OnlyDeleted {
			ctx.PrintPrimaryOutputTableExt(table, true, []int{0, 1, 2, 3, 4, 5, 6})
		} else {
			ctx.PrintPrimaryOutputTableExt(table, true, []int{0, 2, 3, 4, 5, 6})
		}

		return nil

	case cli.OutputFormatText:
		for _, v := range entries {
			ctx.PrintPrimaryOutput("ID:          " + v.ID)
			if v.Deleted {
				ctx.PrintPrimaryOutput("Deleted:     true")
			}
			ctx.PrintPrimaryOutput("Uri:         " + v.URI)
			ctx.PrintPrimaryOutput("Title:       " + v.Title)
			ctx.PrintPrimaryOutput(fmt.Sprintf("Visits (%d):", len(v.Visits)))
			for _, visit := range v.Visits {
				ctx.PrintPrimaryOutput(fmt.Sprintf("  - %s (%s)", visit.VisitDate.Format(ctx.Opt.TimeFormat), visit.TransitionType.ConstantString()))
			}
			ctx.PrintPrimaryOutput("")
		}
		return nil

	case cli.OutputFormatJson:
		json := langext.A{}
		for _, v := range entries {
			json = append(json, v.ToJSON(ctx))
		}
		ctx.PrintPrimaryOutputJSON(json)
		return nil

	case cli.OutputFormatXML:
		type xmlroot struct {
			Entries []any
			XMLName struct{} `xml:"history"`
		}
		node := xmlroot{Entries: make([]any, 0, len(entries))}
		for _, v := range entries {
			node.Entries = append(node.Entries, v.ToSingleXML(ctx, a.IncludeDeleted))
		}
		ctx.PrintPrimaryOutputXML(node)
		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
