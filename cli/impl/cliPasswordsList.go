package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"strconv"
	"time"
)

type CLIArgumentsPasswordsList struct {
	ShowPasswords      bool
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time
	IncludeDeleted     bool
	OnlyDeleted        bool

	CLIArgumentsPasswordsUtil
}

func NewCLIArgumentsPasswordsList() *CLIArgumentsPasswordsList {
	return &CLIArgumentsPasswordsList{
		ShowPasswords:      false,
		IgnoreSchemaErrors: false,
		Sort:               nil,
		Limit:              nil,
		Offset:             nil,
		After:              nil,
		IncludeDeleted:     false,
		OnlyDeleted:        false,
	}
}

func (a *CLIArgumentsPasswordsList) Mode() cli.Mode {
	return cli.ModePasswordsList
}

func (a *CLIArgumentsPasswordsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsPasswordsList) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML, cli.OutputFormatCSV, cli.OutputFormatTSV}
}

func (a *CLIArgumentsPasswordsList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient passwords list", "List passwords"},
		{"          [--show-passwords]", "Show the actual passwords"},
		{"          [--ignore-schema-errors]", "Skip records that cannot be decoded into a password schema"},
		{"          [--after <rfc3339>]", "Return only fields updated after this date"},
		{"          [--sort <sort>]", "Sort the result by (newest|index|oldest)"},
		{"          [--limit <n>]", "Return max <n> elements"},
		{"          [--offset <o>]", "Skip the first <n> elements"},
		{"          [--include-deleted]", "Show deleted entries"},
		{"          [--only-deleted]", "Show only deleted entries"},
	}
}

func (a *CLIArgumentsPasswordsList) FullHelp() []string {
	return []string{
		"$> ffsclient passwords list [--show-passwords] [--ignore-schema-errors] [--after <rfc3339>] [--sort <sort>] [--limit <n>] [--offset <o>] [--include-deleted] [--only-deleted]",
		"",
		"List passwords",
		"",
		"Does not show passwords by default. Use --show-passwords to output them.",
		"If --ignore-schema-errors is not supplied the programm returns with exitcode [0] if any record in the passwords collection has invalid data. Otherwise we simply skip that record.",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned.",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest ).",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
		"By default we skip entries with {deleted:true}, this can be changed with --include-deleted and --only-deleted.",
	}
}

func (a *CLIArgumentsPasswordsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "show-passwords" && arg.Value == nil {
			a.ShowPasswords = true
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

func (a *CLIArgumentsPasswordsList) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[List Passwords]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	records, err := client.ListRecords(ctx, session, consts.CollectionPasswords, a.After, a.Sort, false, true, a.Limit, a.Offset)
	if err != nil {
		return err
	}

	passwords, err := models.UnmarshalPasswords(ctx, records, a.IgnoreSchemaErrors)
	if err != nil {
		return err
	}

	// ========================================================================

	return a.printOutput(ctx, passwords)
}

func (a *CLIArgumentsPasswordsList) printOutput(ctx *cli.FFSContext, passwords []models.PasswordRecord) error {
	passwords = a.filterDeleted(ctx, passwords, a.IncludeDeleted, a.OnlyDeleted)

	ofmt := langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable)
	switch ofmt {

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(passwords))
		table = append(table, []string{"ID", "DELETED", "HOST", "USERNAME", "PASSWORD"})
		for _, v := range passwords {
			table = append(table, []string{
				v.ID,
				langext.FormatBool(v.Deleted, "true", "false"),
				v.Hostname,
				v.Username,
				v.FormatPassword(a.ShowPasswords),
			})
		}

		if a.IncludeDeleted && !a.OnlyDeleted {
			ctx.PrintPrimaryOutputTableExt(table, []int{0, 1, 2, 3, 4})
		} else {
			ctx.PrintPrimaryOutputTableExt(table, []int{0, 2, 3, 4})
		}

		return nil

	case cli.OutputFormatText:
		for _, v := range passwords {
			if schema := urlSchemaRegex.FindString(v.Hostname); schema != "" {
				ctx.PrintPrimaryOutput(schema + v.Username + ":" + v.FormatPassword(a.ShowPasswords) + "@" + v.Hostname[len(schema):])
			} else {
				ctx.PrintPrimaryOutput(v.Username + ":" + v.Password + "@" + v.Hostname)
			}
		}
		return nil

	case cli.OutputFormatJson:
		arr := langext.A{}
		for _, v := range passwords {
			arr = append(arr, v.ToJSON(ctx, a.ShowPasswords))
		}
		ctx.PrintPrimaryOutputJSON(arr)
		return nil

	case cli.OutputFormatXML:
		type xml struct {
			Entries []any
			XMLName struct{} `xml:"Passwords"`
		}
		node := xml{Entries: make([]any, 0, len(passwords))}
		for _, v := range passwords {
			node.Entries = append(node.Entries, v.ToXML(ctx, "Password", a.ShowPasswords))
		}
		ctx.PrintPrimaryOutputXML(node)
		return nil

	case cli.OutputFormatTSV:
		fallthrough
	case cli.OutputFormatCSV:
		table := make([][]string, 0, len(passwords))
		table = append(table, []string{"ID", "Deleted", "Hostname", "Username", "Password", "FormSubmitUrl", "PasswordField", "UsernameField", "Created", "HTTPRealm", "LastUsed", "PasswordChanged", "TimesUsed"})
		for _, v := range passwords {
			table = append(table, []string{
				v.ID,
				langext.FormatBool(v.Deleted, "true", "false"),
				v.Hostname,
				v.Username,
				v.FormatPassword(a.ShowPasswords),
				v.FormSubmitURL,
				v.PasswordField,
				v.UsernameField,
				fmtOptDate(ctx, v.Created),
				langext.Coalesce(v.HTTPRealm, ""),
				fmtOptDate(ctx, v.LastUsed),
				fmtOptDate(ctx, v.PasswordChanged),
				fmt.Sprintf("%d", langext.Coalesce(v.TimesUsed, 0)),
			})
		}

		ctx.PrintPrimaryOutputCSV(table, ofmt == cli.OutputFormatTSV)

		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
