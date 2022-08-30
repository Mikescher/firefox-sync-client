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

type CLIArgumentsPasswordsGet struct {
	Query            string
	QueryIsHost      bool
	QueryIsExactHost bool
	QueryIsID        bool
}

func NewCLIArgumentsPasswordsGet() *CLIArgumentsPasswordsGet {
	return &CLIArgumentsPasswordsGet{
		Query:            "",
		QueryIsHost:      false,
		QueryIsExactHost: false,
		QueryIsID:        false,
	}
}

func (a *CLIArgumentsPasswordsGet) Mode() cli.Mode {
	return cli.ModePasswordsGet
}

func (a *CLIArgumentsPasswordsGet) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsPasswordsGet) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient passwords get <host|id>", "Insert a new password"},
		{"          [--host | --exact-host | --id]", "Specify that the supplier argument is a host / record-id (otherwise both is possible)"},
	}
}

func (a *CLIArgumentsPasswordsGet) FullHelp() []string {
	return []string{
		"$> ffsclient passwords get <host|id> [--host | --exact-host | --id]",
		"",
		"Get a password",
		"",
		"By default we can supply a host or a record-id.",
		"If --host is specified, the query is parsed as an URI and we return the password that matches the host.",
		"If --exact-host is specified, the query is matched exactly against the host field in the password record.",
		"If --id is specified, the query is matched exactly agains the record-id.",
		"If --id is _not_ specified this method needs to query all passwords from the server and do a local search.",
	}
}

func (a *CLIArgumentsPasswordsGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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

func (a *CLIArgumentsPasswordsGet) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Get Password]")
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

	ctx.PrintPrimaryOutput(record.Password)
	return 0
}

func (a *CLIArgumentsPasswordsGet) getRecord(ctx *cli.FFSContext, client *syncclient.FxAClient, session syncclient.FFSyncSession) (models.PasswordRecord, bool, error) {

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

func (a *CLIArgumentsPasswordsGet) extUrlParse(v string) (*url.URL, error) {

	if !urlSchemaRegex.MatchString(v) {
		v = "generic://" + v
	}

	return url.Parse(v)
}
