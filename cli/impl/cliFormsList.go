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

type CLIArgumentsFormsList struct {
	NameFilter         *[]string
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time
	IncludeDeleted     bool
	OnlyDeleted        bool

	CLIArgumentsFormsUtil
}

func NewCLIArgumentsFormsList() *CLIArgumentsFormsList {
	return &CLIArgumentsFormsList{
		NameFilter:         nil,
		IgnoreSchemaErrors: false,
		Sort:               nil,
		Limit:              nil,
		Offset:             nil,
		After:              nil,
		IncludeDeleted:     false,
		OnlyDeleted:        false,

		CLIArgumentsFormsUtil: CLIArgumentsFormsUtil{},
	}
}

func (a *CLIArgumentsFormsList) Mode() cli.Mode {
	return cli.ModeFormsList
}

func (a *CLIArgumentsFormsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsFormsList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient forms list", "List form autocomplete suggestions"},
		{"          [--name <n>]", "Show only entries with the specified name"},
		{"          [--ignore-schema-errors]", "Skip records that cannot be decoded into a password schema"},
		{"          [--after <rfc3339>]", "Return only fields updated after this date"},
		{"          [--sort <sort>]", "Sort the result by (newest|index|oldest)"},
		{"          [--limit <n>]", "Return max <n> elements"},
		{"          [--offset <o>]", "Skip the first <n> elements"},
		{"          [--include-deleted]", "Show deleted entries"},
		{"          [--only-deleted]", "Show only deleted entries"},
	}
}

func (a *CLIArgumentsFormsList) FullHelp() []string {
	return []string{
		"$> ffsclient forms list [--name <n>] [--ignore-schema-errors] [--after <rfc3339>] [--sort <sort>] [--limit <n>] [--offset <o>] [--include-deleted] [--only-deleted]",
		"",
		"List HTML-Form autocomplete suggestions",
		"",
		"If --ignore-schema-errors is not supplied the programm returns with exitcode [0] if any record in the forms collection has invalid data. Otherwise we simply skip that record.",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned.",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest ).",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
		"By default we skip entries with {deleted:true}, this can be changed with --include-deleted and --only-deleted.",
		"",
		"You can filter the returned entries types with --name, the name-filter can be specified multiple times to filter multiple names",
	}
}

func (a *CLIArgumentsFormsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "name" && arg.Value != nil {
			if a.NameFilter == nil {
				a.NameFilter = &[]string{*arg.Value}
			} else {
				v := append(*a.NameFilter, *arg.Value)
				a.NameFilter = &v
			}
			continue
		}
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

func (a *CLIArgumentsFormsList) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[List Forms]")
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

	records, err := client.ListRecords(ctx, session, consts.CollectionForms, a.After, a.Sort, false, true, a.Limit, a.Offset)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	forms, err := models.UnmarshalForms(ctx, records, a.IgnoreSchemaErrors)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	return a.printOutput(ctx, forms)
}

func (a *CLIArgumentsFormsList) printOutput(ctx *cli.FFSContext, forms []models.FormRecord) int {
	forms = a.filterDeleted(ctx, forms, a.IncludeDeleted, a.OnlyDeleted, a.NameFilter)

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable) {

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(forms))
		table = append(table, []string{"ID", "DELETED", "DATE", "NAME", "VALUE"})
		for _, v := range forms {
			table = append(table, []string{
				v.ID,
				langext.FormatBool(v.Deleted, "true", "false"),
				v.LastModified.Format(ctx.Opt.TimeFormat),
				v.Name,
				v.Value,
			})
		}

		if a.IncludeDeleted && !a.OnlyDeleted {
			ctx.PrintPrimaryOutputTableExt(table, true, []int{0, 1, 2, 3, 4})
		} else {
			ctx.PrintPrimaryOutputTableExt(table, true, []int{0, 2, 3, 4})
		}

		return 0

	case cli.OutputFormatText:
		for _, v := range forms {
			ctx.PrintPrimaryOutput("ID:          " + v.ID)
			if v.Deleted {
				ctx.PrintPrimaryOutput("Deleted:     true")
			}
			ctx.PrintPrimaryOutput("Date:        " + v.LastModified.Format(ctx.Opt.TimeFormat))
			ctx.PrintPrimaryOutput("Name:        " + v.Name)
			ctx.PrintPrimaryOutput("Value:       " + v.Value)
			ctx.PrintPrimaryOutput("")
		}
		return 0

	case cli.OutputFormatJson:
		json := langext.A{}
		for _, v := range forms {
			json = append(json, v.ToJSON(ctx))
		}
		ctx.PrintPrimaryOutputJSON(json)
		return 0

	case cli.OutputFormatXML:
		type xmlroot struct {
			Entries []any
			XMLName struct{} `xml:"forms"`
		}
		node := xmlroot{Entries: make([]any, 0, len(forms))}
		for _, v := range forms {
			node.Entries = append(node.Entries, v.ToSingleXML(ctx, a.IncludeDeleted))
		}
		ctx.PrintPrimaryOutputXML(node)
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}
