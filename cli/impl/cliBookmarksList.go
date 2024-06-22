package impl

import (
	"encoding/xml"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"ffsyncclient/netscapefmt"
	"fmt"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"strconv"
	"strings"
	"time"
)

type CLIArgumentsBookmarksList struct {
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time
	IncludeDeleted     bool
	OnlyDeleted        bool
	TypeFilter         *[]models.BookmarkType
	ParentFilter       *[]string
	LinearOutput       bool

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksList() *CLIArgumentsBookmarksList {
	return &CLIArgumentsBookmarksList{
		IgnoreSchemaErrors: false,
		Sort:               nil,
		Limit:              nil,
		Offset:             nil,
		After:              nil,
		IncludeDeleted:     false,
		OnlyDeleted:        false,
		TypeFilter:         nil,
		LinearOutput:       false,
	}
}

func (a *CLIArgumentsBookmarksList) Mode() cli.Mode {
	return cli.ModeBookmarksList
}

func (a *CLIArgumentsBookmarksList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsBookmarksList) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML, cli.OutputFormatNetscape, cli.OutputFormatTSV, cli.OutputFormatCSV}
}

func (a *CLIArgumentsBookmarksList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient bookmarks list", "List bookmarks (use --format to define the format)"},
		{"          [--ignore-schema-errors]", "Skip records that cannot be decoded into a bookmark schema"},
		{"          [--after <rfc3339>]", "Return only fields updated after this date"},
		{"          [--sort <sort>]", "Sort the result by (newest|index|oldest)"},
		{"          [--limit <n>]", "Return max <n> elements"},
		{"          [--offset <o>]", "Skip the first <n> elements"},
		{"          [--include-deleted]", "Show deleted entries"},
		{"          [--only-deleted]", "Show only deleted entries"},
		{"          [--type <folder|separator|bookmark|...>]", "Show only entries with the specified type"},
		{"          [--parent <id>]", "Show only entries with the specified parent (by record-id), can be specified multiple times"},
		{"          [--linear", "Do not output the folder hierachy"},
	}
}

func (a *CLIArgumentsBookmarksList) FullHelp() []string {
	return []string{
		"$> ffsclient bookmarks list [--ignore-schema-errors] [--after <rfc3339>] [--sort <sort>] [--limit <n>] [--offset <o>] [--include-deleted] [--only-deleted] [--type <bmt>] [--linear]",
		"",
		"List bookmarks",
		"",
		"If --ignore-schema-errors is not supplied the programm returns with exitcode [0] if any record in the bookmarks collection has invalid data. Otherwise we simply skip that record.",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned.",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest ).",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
		"By default we skip entries with {deleted:true}, this can be changed with --include-deleted and --only-deleted.",
		"If --linear is not supplied the output will (depending on the format) print the bookmarks in their folder hierachy, wiith --linear the data is printed as a flat array",
		"",
		"The following --format output-formats are possible:",
		"  * [--format text]     Simple text output",
		"  * [--format table]    Simple tabular output",
		"  * [--format json]     Output bookmark data as json",
		"  * [--format netscape] Output bookmark data as netscape bookmarks html (same as the firefox bookmarks.html format)",
		"  * [--format xml]      Output bookmark data as XML",
		"",
		"You can filter the returned bookmark types with --type, the following types are possible:",
		"(Specify multiple types by having multiple --type parameter)",
		"  * [--type bookmark]",
		"  * [--type microsummary]    (deprecated)",
		"  * [--type query]",
		"  * [--type folder]",
		"  * [--type livemark]",
		"  * [--type separator]",
		"You can also filter the returned bookmarks by their parent with --parent (needs a record-id).",
		"This can also be used specify multiple parents with multiple --parent arguments.",
	}
}

func (a *CLIArgumentsBookmarksList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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
		if arg.Key == "linear" && arg.Value == nil {
			a.LinearOutput = true
			continue
		}
		if arg.Key == "type" && arg.Value != nil {
			if a.TypeFilter == nil {
				a.TypeFilter = &[]models.BookmarkType{models.BookmarkType(*arg.Value)}
			} else {
				v := append(*a.TypeFilter, models.BookmarkType(*arg.Value))
				a.TypeFilter = &v
			}
			continue
		}
		if arg.Key == "parent" && arg.Value != nil {
			if a.ParentFilter == nil {
				a.ParentFilter = &[]string{*arg.Value}
			} else {
				v := append(*a.ParentFilter, *arg.Value)
				a.ParentFilter = &v
			}
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

func (a *CLIArgumentsBookmarksList) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[List Bookmarks]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	records, err := client.ListRecords(ctx, session, consts.CollectionBookmarks, a.After, a.Sort, false, true, a.Limit, a.Offset)
	if err != nil {
		return err
	}

	bookmarks, err := models.UnmarshalBookmarks(ctx, records, a.IgnoreSchemaErrors)
	if err != nil {
		return err
	}

	// ========================================================================

	return a.printOutput(ctx, bookmarks)
}

func (a *CLIArgumentsBookmarksList) printOutput(ctx *cli.FFSContext, bookmarks []models.BookmarkRecord) error {

	parentFilter := a.ParentFilter
	ofmt := langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable)

	includeParent := parentFilter != nil && !a.LinearOutput && (ofmt == cli.OutputFormatJson || ofmt == cli.OutputFormatXML)

	bookmarks = a.filterDeleted(ctx, bookmarks, a.IncludeDeleted, a.OnlyDeleted, a.TypeFilter, parentFilter, includeParent)

	switch ofmt {

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(bookmarks))
		table = append(table, []string{"ID", "TYPE", "DELETED", "TITLE", "URI"})
		for _, v := range bookmarks {
			table = append(table, []string{
				v.ID,
				string(v.Type),
				langext.FormatBool(v.Deleted, "true", "false"),
				v.Title,
				v.URI,
			})
		}

		if a.IncludeDeleted && !a.OnlyDeleted {
			ctx.PrintPrimaryOutputTableExt(table, []int{0, 1, 2, 3, 4})
		} else {
			ctx.PrintPrimaryOutputTableExt(table, []int{0, 1, 3, 4})
		}

		return nil

	case cli.OutputFormatText:
		for _, v := range bookmarks {
			ctx.PrintPrimaryOutput("ID:          " + v.ID)
			if v.Deleted {
				ctx.PrintPrimaryOutput("Deleted:     true")
			}
			ctx.PrintPrimaryOutput("Type:        " + string(v.Type))
			ctx.PrintPrimaryOutput("Title:       " + v.Title)
			if v.Description != "" {
				ctx.PrintPrimaryOutput("Description: " + strings.ReplaceAll(strings.ReplaceAll(v.Description, "\r", ""), "\n", " "))
			}
			if v.URI != "" {
				ctx.PrintPrimaryOutput("URI:         " + v.URI)
			}
			if v.SiteURI != "" {
				ctx.PrintPrimaryOutput("SiteURI:     " + v.URI)
			}
			if v.FeedURI != "" {
				ctx.PrintPrimaryOutput("FeedURI:     " + v.URI)
			}
			if v.Type == models.BookmarkTypeFolder || v.Type == models.BookmarkTypeLivemark {
				if len(v.Children) > 0 {
					ctx.PrintPrimaryOutput("Children:    " + "['" + strings.Join(v.Children, "', '") + "']")
				} else {
					ctx.PrintPrimaryOutput("Children:    " + "[]")
				}
			}
			ctx.PrintPrimaryOutput("")
		}
		return nil

	case cli.OutputFormatJson:
		if a.LinearOutput {
			json := langext.A{}
			for _, v := range bookmarks {
				json = append(json, v.ToJSON(ctx))
			}
			ctx.PrintPrimaryOutputJSON(json)
			return nil
		} else {
			roots, unreferenced, missing := a.calculateTree(ctx, bookmarks, langext.Coalesce(parentFilter, nil))
			jsonRoots := langext.H{}
			for _, v := range roots {
				jsonRoots[v.ID] = v.ToTreeJSON(ctx)
			}
			jsonUnreferenced := langext.A{}
			for _, v := range unreferenced {
				jsonUnreferenced = append(jsonUnreferenced, v.ToJSON(ctx))
			}

			ctx.PrintPrimaryOutputJSON(langext.H{"bookmarks": jsonRoots, "unreferenced": jsonUnreferenced, "missing": langext.ForceArray(missing)})
			return nil
		}

	case cli.OutputFormatXML:
		if a.LinearOutput {
			type xmlroot struct {
				Entries []any
				XMLName struct{} `xml:"Bookmarks"`
			}
			node := xmlroot{Entries: make([]any, 0, len(bookmarks))}
			for _, v := range bookmarks {
				node.Entries = append(node.Entries, v.ToSingleXML(ctx, a.IncludeDeleted))
			}
			ctx.PrintPrimaryOutputXML(node)
			return nil
		} else {
			roots, unreferenced, missing := a.calculateTree(ctx, bookmarks, langext.Coalesce(parentFilter, nil))
			type xmlroot struct {
				Entries []any
				XMLName struct{} `xml:"Bookmarks"`
				Missing string   `xml:"Missing,attr,omitempty"`
			}
			node := xmlroot{Entries: make([]any, 0, len(bookmarks)), Missing: strings.Join(missing, ", ")}
			for _, v := range roots {
				node.Entries = append(node.Entries, v.ToTreeXML(ctx, a.IncludeDeleted))
			}
			if len(unreferenced) > 0 {
				type xmlentry struct {
					XMLName xml.Name
					Entries []any
				}
				e := make([]any, 0)
				for _, v := range unreferenced {
					e = append(e, v.ToSingleXML(ctx, a.IncludeDeleted))
				}
				node.Entries = append(node.Entries, xmlentry{
					XMLName: xml.Name{Local: "@unreferenced"},
					Entries: e,
				})
			}
			ctx.PrintPrimaryOutputXML(node)
			return nil
		}

	case cli.OutputFormatNetscape:
		roots, _, _ := a.calculateTree(ctx, bookmarks, nil)
		nc := netscapefmt.Format(ctx, roots)
		ctx.PrintPrimaryOutput(nc)
		return nil

	case cli.OutputFormatTSV:
		fallthrough
	case cli.OutputFormatCSV:
		table := make([][]string, 0, len(bookmarks))
		table = append(table, []string{"ID", "ParentID", "Type", "Deleted", "Title", "URI", "Keyword", "DateAdded", "Description", "Tags"})
		for _, v := range bookmarks {
			table = append(table, []string{
				v.ID,
				v.ParentID,
				string(v.Type),
				langext.FormatBool(v.Deleted, "true", "false"),
				v.Title,
				v.URI,
				v.Keyword,
				fmtOptDate(ctx, v.DateAdded),
				v.Description,
				strings.Join(v.Tags, ";"),
			})
		}

		ctx.PrintPrimaryOutputCSV(table, ofmt == cli.OutputFormatTSV)

		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
