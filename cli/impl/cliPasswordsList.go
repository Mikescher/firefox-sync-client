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

type CLIArgumentsPasswordsList struct {
	ShowPasswords      bool
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time
}

func NewCLIArgumentsPasswordsList() *CLIArgumentsPasswordsList {
	return &CLIArgumentsPasswordsList{
		ShowPasswords:      false,
		IgnoreSchemaErrors: false,
		Sort:               nil,
		Limit:              nil,
		Offset:             nil,
		After:              nil,
	}
}

func (a *CLIArgumentsPasswordsList) Mode() cli.Mode {
	return cli.ModePasswordsList
}

func (a *CLIArgumentsPasswordsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
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
	}
}

func (a *CLIArgumentsPasswordsList) FullHelp() []string {
	return []string{
		"$> ffsclient passwords list [--show-passwords] [--ignore-schema-errors] [--after <rfc3339>] [--sort <sort>] [--limit <n>] [--offset <o>]",
		"",
		"List passwords",
		"",
		"Does not show passwords by default. Use --show-passwords to output them.",
		"If --ignore-schema-errors is not supplied the programm returns with an exitcode <> 0 if any record in the passwords collection has invalid data. Otherwise we simply skip that record.",
		"If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned.",
		"If --sort is specified the resulting records are sorted by ( newest | index | oldest ).",
		"The --limit and --offset parameter can be used to get a subset of the result and paginate through it.",
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
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsList) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[List Passwords]")
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

	records, err := client.ListRecords(ctx, session, consts.CollectionPasswords, a.After, a.Sort, false, true, a.Limit, a.Offset)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	passwords, err := models.ParsePasswords(ctx, records, a.IgnoreSchemaErrors)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	return a.printOutput(ctx, passwords)
}

func (a *CLIArgumentsPasswordsList) printOutput(ctx *cli.FFSContext, passwords []models.PasswordRecord) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable) {

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(passwords))
		table = append(table, []string{"ID", "HOST", "USERNAME", "PASSWORD"})
		for _, v := range passwords {
			table = append(table, []string{
				v.ID,
				v.Hostname,
				v.Username,
				v.FormatPassword(a.ShowPasswords),
			})
		}

		ctx.PrintPrimaryOutputTable(table, true)
		return 0

	case cli.OutputFormatText:
		for _, v := range passwords {
			if schema := urlSchemaRegex.FindString(v.Hostname); schema != "" {
				ctx.PrintPrimaryOutput(schema + v.Username + ":" + v.FormatPassword(a.ShowPasswords) + "@" + v.Hostname[len(schema):])
			} else {
				ctx.PrintPrimaryOutput(v.Username + ":" + v.Password + "@" + v.Hostname)
			}
		}
		return 0

	case cli.OutputFormatJson:
		arr := langext.A{}
		for _, v := range passwords {
			arr = append(arr, v.ToJSON(ctx, a.ShowPasswords))
		}
		ctx.PrintPrimaryOutputJSON(arr)
		return 0

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
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}