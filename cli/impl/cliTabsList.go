package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"strconv"
)

type CLIArgumentsTabsList struct {
	ClientFilter       *[]string
	IgnoreSchemaErrors bool
	Limit              *int
	Offset             *int
	IncludeDeleted     bool
	OnlyDeleted        bool

	CLIArgumentsTabsUtil
}

func NewCLIArgumentsTabsList() *CLIArgumentsTabsList {
	return &CLIArgumentsTabsList{
		ClientFilter:       nil,
		IgnoreSchemaErrors: false,
		Limit:              nil,
		Offset:             nil,
		IncludeDeleted:     false,
		OnlyDeleted:        false,
	}
}

func (a *CLIArgumentsTabsList) Mode() cli.Mode {
	return cli.ModeTabsList
}

func (a *CLIArgumentsTabsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsTabsList) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML, cli.OutputFormatCSV, cli.OutputFormatTSV}
}

func (a *CLIArgumentsTabsList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient tabs list", "List synchronized tabs"},
		{"          [--client <n>]", "Show only entries from the specified client (must be a valid client-id)"},
		{"          [--ignore-schema-errors]", "Skip records that cannot be decoded into a tab schema"},
		{"          [--limit <n>]", "Return max <n> elements (clients)"},
		{"          [--offset <o>]", "Skip the first <n> elements (clients)"},
		{"          [--include-deleted]", "Show deleted entries"},
		{"          [--only-deleted]", "Show only deleted entries"},
	}
}

func (a *CLIArgumentsTabsList) FullHelp() []string {
	return []string{
		"$> ffsclient tabs list [--client <n>] [--ignore-schema-errors] [--limit <n>] [--offset <o>] [--include-deleted] [--only-deleted]",
		"",
		"List synchronized tabs from clients",
		"",
		"If --ignore-schema-errors is not supplied the programm returns with exitcode [60] if any record in the tabs collection has invalid data. Otherwise we simply skip that record.",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
		"But because the underlying records are grouped together by client, this limits the number of returned clients and not directly the number of returned records",
		"By default we skip entries with {deleted:true}, this can be changed with --include-deleted and --only-deleted.",
		"",
		"You can filter the returned entries types with --client, the client-filter can be specified multiple times to show tabs of multiple clients. The value must be a valid client-id.",
		"",
		"Because the --client filtering is done client-side it is possible that results on later pages are not shown. They must explicitly be requested with an --offset value.",
	}
}

func (a *CLIArgumentsTabsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "client" && arg.Value != nil {
			if a.ClientFilter == nil {
				a.ClientFilter = &[]string{*arg.Value}
			} else {
				v := append(*a.ClientFilter, *arg.Value)
				a.ClientFilter = &v
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
		if arg.Key == "limit" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Limit = langext.Ptr(int(v))
				continue
			}
			return fferr.DirectOutput.New(fmt.Sprintf("Failed to parse number argument '--%s': '%s'", arg.Key, *arg.Value))
		}
		if arg.Key == "offset" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Offset = langext.Ptr(int(v))
				continue
			}
			return fferr.DirectOutput.New(fmt.Sprintf("Failed to parse number argument '--%s': '%s'", arg.Key, *arg.Value))
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsTabsList) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[List Forms]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	records, err := client.ListRecords(ctx, session, consts.CollectionTabs, nil, nil, false, true, a.Limit, a.Offset)
	if err != nil {
		return err
	}

	ctabs, err := models.UnmarshalTabs(ctx, records, a.IgnoreSchemaErrors)
	if err != nil {
		return err
	}

	// ========================================================================

	return a.printOutput(ctx, ctabs)
}

func (a *CLIArgumentsTabsList) printOutput(ctx *cli.FFSContext, indata []models.TabClientRecord) error {
	alltabs := a.filterDeletedSingle(ctx, langext.ArrFlatten(indata, func(v models.TabClientRecord) []models.TabRecord { return v.Tabs }), a.IncludeDeleted, a.OnlyDeleted, a.ClientFilter)
	clienttabs := a.filterDeletedMulti(ctx, indata, a.IncludeDeleted, a.OnlyDeleted, a.ClientFilter)

	ofmt := langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable)
	switch ofmt {

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(alltabs))
		table = append(table, []string{"CLIENTID", "CLIENT", "DELETED", "INDEX", "TITLE", "URL", "HISTORY"})
		for _, v := range alltabs {
			table = append(table, []string{
				v.ClientID,
				v.ClientName,
				langext.FormatBool(v.ClientDeleted, "true", "false"),
				fmt.Sprintf("%d", v.Index),
				v.Title,
				a.LastHistory(v, ""),
				fmt.Sprintf("%d", len(v.UrlHistory)),
			})
		}

		viscols := make([]int, 0, 16)
		if a.ClientFilter == nil || len(*a.ClientFilter) != 1 {
			viscols = append(viscols, 0, 1)
		}
		if a.IncludeDeleted && !a.OnlyDeleted {
			viscols = append(viscols, 2)
		}
		viscols = append(viscols, 3, 4, 5, 6)

		ctx.PrintPrimaryOutputTableExt(table, viscols)

		return nil

	case cli.OutputFormatText:
		for _, c := range clienttabs {

			ctx.PrintPrimaryOutput(c.Name)
			if !c.Deleted {
				ctx.PrintPrimaryOutput("[ID]: " + c.ID)
			} else {
				ctx.PrintPrimaryOutput("[ID]: " + c.ID + " (Deleted)")
			}
			ctx.PrintPrimaryOutput("============================================================")
			ctx.PrintPrimaryOutput("")

			for _, v := range c.Tabs {

				ctx.PrintPrimaryOutput("Index:       " + strconv.Itoa(v.Index))
				ctx.PrintPrimaryOutput("Title:       " + v.Title)
				ctx.PrintPrimaryOutput("Icon:        " + v.Icon)
				ctx.PrintPrimaryOutput("URL:         " + a.LastHistory(v, ""))
				ctx.PrintPrimaryOutput("LastUsed:    " + v.LastUsed.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat))
				ctx.PrintPrimaryOutput("")
			}

			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("")

		}
		return nil

	case cli.OutputFormatJson:
		json := langext.A{}
		for _, v := range alltabs {
			json = append(json, v.ToJSON(ctx))
		}
		ctx.PrintPrimaryOutputJSON(json)
		return nil

	case cli.OutputFormatXML:
		type xmlroot struct {
			Entries []any
			XMLName struct{} `xml:"tabs"`
		}
		node := xmlroot{Entries: make([]any, 0, len(clienttabs))}
		for _, v := range clienttabs {
			node.Entries = append(node.Entries, v.ToSingleXML(ctx, a.IncludeDeleted))
		}
		ctx.PrintPrimaryOutputXML(node)
		return nil

	case cli.OutputFormatTSV:
		fallthrough
	case cli.OutputFormatCSV:
		table := make([][]string, 0, len(alltabs))
		table = append(table, []string{"ClientID", "ClientName", "ClientDeleted", "Index", "Title", "URL", "Icon", "HistoryCount", "LastUsed"})
		for _, v := range alltabs {
			table = append(table, []string{
				v.ClientID,
				v.ClientName,
				langext.FormatBool(v.ClientDeleted, "true", "false"),
				fmt.Sprintf("%d", v.Index),
				v.Title,
				a.LastHistory(v, ""),
				v.Icon,
				fmt.Sprintf("%d", len(v.UrlHistory)),
				v.LastUsed.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
			})
		}

		ctx.PrintPrimaryOutputCSV(table, ofmt == cli.OutputFormatTSV)

		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
