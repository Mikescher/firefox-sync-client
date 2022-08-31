package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"net/url"
	"strings"
)

type CLIArgumentsPasswordsDelete struct {
	Query            string
	QueryIsHost      bool
	QueryIsExactHost bool
	QueryIsID        bool
}

func NewCLIArgumentsPasswordsDelete() *CLIArgumentsPasswordsDelete {
	return &CLIArgumentsPasswordsDelete{
		Query:            "",
		QueryIsHost:      false,
		QueryIsExactHost: false,
		QueryIsID:        false,
	}
}

func (a *CLIArgumentsPasswordsDelete) Mode() cli.Mode {
	return cli.ModePasswordsDelete
}

func (a *CLIArgumentsPasswordsDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsPasswordsDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient passwords delete <host|id>", "Delete a single password"},
		{"          [--host | --exact-host | --id]", "Specify that the supplied argument is a host / record-id (otherwise both is possible)"},
	}
}

func (a *CLIArgumentsPasswordsDelete) FullHelp() []string {
	return []string{
		"$> ffsclient passwords delete <host|id> [--host | --exact-host | --id]",
		"",
		"Delete a single password",
		"",
		"By default we can supply a host or a record-id.",
		"If --host is specified, the query is parsed as an URI and we return the password that matches the host.",
		"If --exact-host is specified, the query is matched exactly against the host field in the password record.",
		"If --id is specified, the query is matched exactly agains the record-id.",
		"If --id is _not_ specified this method needs to query all passwords from the server and do a local search.",
		"If multiple passwords match the first match is deleted.",
		"If no matching password is found the exitcode [82] is returned",
	}
}

func (a *CLIArgumentsPasswordsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Query = positionalArgs[0]

	for _, arg := range optionArgs {
		if arg.Key == "host" && arg.Value == nil {
			a.QueryIsHost = true
			continue
		}
		if arg.Key == "exact-host" && arg.Value == nil {
			a.QueryIsExactHost = true
			continue
		}
		if arg.Key == "id" && arg.Value == nil {
			a.QueryIsID = true
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsDelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete Password]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Query", a.Query)

	if langext.BoolCount(a.QueryIsID, a.QueryIsExactHost, a.QueryIsHost) > 1 {
		ctx.PrintFatalMessage("Must specify at most one of --id, --exact-host, --host")
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

	record, found, err := a.getRecord(ctx, client, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !found {
		ctx.PrintErrorMessage("Record not found")
		return consts.ExitcodePasswordNotFound
	}

	ctx.PrintVerbose("Delete Record " + record.ID)

	err = client.DeleteRecord(ctx, session, consts.CollectionPasswords, record.ID)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	return a.printOutput(ctx, record)
}

func (a *CLIArgumentsPasswordsDelete) getRecord(ctx *cli.FFSContext, client *syncclient.FxAClient, session syncclient.FFSyncSession) (models.PasswordRecord, bool, error) {

	// #### VARIANT 1: <QueryIsID>

	if a.QueryIsID {
		ctx.PrintVerbose("Query record directly by ID")

		record, err := client.GetRecord(ctx, session, consts.CollectionPasswords, a.Query, true)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return models.PasswordRecord{}, false, nil
		}
		if err != nil {
			return models.PasswordRecord{}, false, errorx.Decorate(err, "failed to query record")
		}

		pwrec, err := models.ParsePassword(ctx, record)
		if err != nil {
			return models.PasswordRecord{}, false, errorx.Decorate(err, "failed to decode password-record")
		}

		return pwrec, true, nil
	}

	records, err := client.ListRecords(ctx, session, consts.CollectionPasswords, nil, nil, false, true, nil, nil)
	if err != nil {
		return models.PasswordRecord{}, false, errorx.Decorate(err, "failed to list passwords")
	}

	allPasswords, err := models.ParsePasswords(ctx, records, true)
	if err != nil {
		return models.PasswordRecord{}, false, errorx.Decorate(err, "failed to decode passwords")
	}

	var parsedURI *url.URL

	if u, err := a.extUrlParse(a.Query); err == nil {
		ctx.PrintVerbose("Parsed query to uri: '" + u.Host + "'")

		parsedURI = u
	}

	// #### VARIANT 2: <QueryIsHost>

	if a.QueryIsHost {
		ctx.PrintVerbose("Search for record by URI")

		if parsedURI == nil {
			return models.PasswordRecord{}, false, errorx.Decorate(err, "cannot parse supplied argument as an URI")
		}

		for _, v := range allPasswords {
			if recordURI, err := url.Parse(v.Hostname); err == nil {
				if strings.ToLower(recordURI.Host) == strings.ToLower(parsedURI.Host) {
					return v, true, nil
				}
			}
		}
		return models.PasswordRecord{}, false, nil
	}

	// #### VARIANT 3: <QueryIsExactHost>

	if a.QueryIsExactHost {
		ctx.PrintVerbose("Search for record by exact Hostname")

		for _, v := range allPasswords {
			if v.Hostname == a.Query {
				return v, true, nil
			}
		}
		return models.PasswordRecord{}, false, nil
	}

	// #### VARIANT 4: <GUESS>

	{
		ctx.PrintVerbose("Search for record (guess query type)")

		for _, v := range allPasswords {
			if v.ID == a.Query {
				return v, true, nil
			}
			if v.Hostname == a.Query {
				return v, true, nil
			}
			if parsedURI != nil {
				if recordURI, err := url.Parse(v.Hostname); err == nil {
					if strings.ToLower(recordURI.Host) == strings.ToLower(parsedURI.Host) {
						return v, true, nil
					}
				}
			}
		}
		return models.PasswordRecord{}, false, nil
	}

}

func (a *CLIArgumentsPasswordsDelete) extUrlParse(v string) (*url.URL, error) {

	if !urlSchemaRegex.MatchString(v) {
		v = "generic://" + v
	}

	return url.Parse(v)
}

func (a *CLIArgumentsPasswordsDelete) printOutput(ctx *cli.FFSContext, password models.PasswordRecord) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(password.ID)
		return 0

	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(password.ToJSON(ctx, true))
		return 0

	case cli.OutputFormatXML:
		ctx.PrintPrimaryOutputXML(password.ToXML(ctx, "Password", true))
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}