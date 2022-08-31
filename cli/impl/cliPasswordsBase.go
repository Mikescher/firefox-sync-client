package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"net/url"
	"strings"
	"time"
)

type CLIArgumentsPasswordsBase struct {
	ShowPasswords      bool
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time

	CLIArgumentsPasswordsUtil
}

func NewCLIArgumentsPasswordsBase() *CLIArgumentsPasswordsBase {
	return &CLIArgumentsPasswordsBase{
		ShowPasswords: false,
		Sort:          nil,
		Limit:         nil,
		Offset:        nil,
		After:         nil,
	}
}

func (a *CLIArgumentsPasswordsBase) Mode() cli.Mode {
	return cli.ModePasswordsBase
}

func (a *CLIArgumentsPasswordsBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsPasswordsBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsPasswordsBase) FullHelp() []string {
	r := []string{
		"$> ffsclient passwords (list|delete|create|update|get)",
		"======================================================",
		"",
	}
	for _, v := range ListSubcommands(a.Mode()) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsPasswordsBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient passwords must be called with a subcommand (eg `ffsclient passwords list`)")
}

func (a *CLIArgumentsPasswordsBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}

type CLIArgumentsPasswordsUtil struct{}

func (a *CLIArgumentsPasswordsUtil) findPasswordRecord(ctx *cli.FFSContext, client *syncclient.FxAClient, session syncclient.FFSyncSession, query string, queryIsID bool, queryIsHost bool, queryIsExactHost bool) (models.PasswordRecord, bool, error) {

	// #### VARIANT 1: <QueryIsID>

	if queryIsID {
		ctx.PrintVerbose("Query record directly by ID")

		record, err := client.GetRecord(ctx, session, consts.CollectionPasswords, query, true)
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

	if u, err := a.extUrlParse(query); err == nil {
		ctx.PrintVerbose("Parsed query to uri: '" + u.Host + "'")

		parsedURI = u
	}

	// #### VARIANT 2: <QueryIsHost>

	if queryIsHost {
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

	if queryIsExactHost {
		ctx.PrintVerbose("Search for record by exact Hostname")

		for _, v := range allPasswords {
			if v.Hostname == query {
				return v, true, nil
			}
		}
		return models.PasswordRecord{}, false, nil
	}

	// #### VARIANT 4: <GUESS>

	{
		ctx.PrintVerbose("Search for record (guess query type)")

		for _, v := range allPasswords {
			if v.ID == query {
				return v, true, nil
			}
			if v.Hostname == query {
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

func (a *CLIArgumentsPasswordsUtil) extUrlParse(v string) (*url.URL, error) {
	if !urlSchemaRegex.MatchString(v) {
		v = "generic://" + v
	}

	return url.Parse(v)
}
