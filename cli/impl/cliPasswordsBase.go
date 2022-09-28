package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
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
		"",
	}
	for _, v := range ListSubcommands(a.Mode(), true) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
		r = append(r, "")
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

func (a *CLIArgumentsPasswordsUtil) findPasswordRecord(ctx *cli.FFSContext, client *syncclient.FxAClient, session syncclient.FFSyncSession, query string, queryIsID bool, queryIsHost bool, queryIsExactHost bool) (models.Record, models.PasswordRecord, bool, error) {

	// #### VARIANT 1: <QueryIsID>

	if queryIsID {
		ctx.PrintVerbose("Query record directly by ID")

		record, err := client.GetRecord(ctx, session, consts.CollectionPasswords, query, true)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return models.Record{}, models.PasswordRecord{}, false, nil
		}
		if err != nil {
			return models.Record{}, models.PasswordRecord{}, false, errorx.Decorate(err, "failed to query record")
		}

		pwrec, err := models.UnmarshalPassword(ctx, record)
		if err != nil {
			return models.Record{}, models.PasswordRecord{}, false, errorx.Decorate(err, "failed to decode password-record")
		}

		return record, pwrec, true, nil
	}

	records, err := client.ListRecords(ctx, session, consts.CollectionPasswords, nil, nil, false, true, nil, nil)
	if err != nil {
		return models.Record{}, models.PasswordRecord{}, false, errorx.Decorate(err, "failed to list passwords")
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
			return models.Record{}, models.PasswordRecord{}, false, errorx.Decorate(err, "cannot parse supplied argument as an URI")
		}

		for _, rec := range records {
			v, err := models.UnmarshalPassword(ctx, rec)
			if err != nil {
				continue
			}
			if recordURI, err := url.Parse(v.Hostname); err == nil {
				if strings.ToLower(recordURI.Host) == strings.ToLower(parsedURI.Host) {
					return rec, v, true, nil
				}
			}
		}
		return models.Record{}, models.PasswordRecord{}, false, nil
	}

	// #### VARIANT 3: <QueryIsExactHost>

	if queryIsExactHost {
		ctx.PrintVerbose("Search for record by exact Hostname")

		for _, rec := range records {
			v, err := models.UnmarshalPassword(ctx, rec)
			if err != nil {
				continue
			}
			if v.Hostname == query {
				return rec, v, true, nil
			}
		}
		return models.Record{}, models.PasswordRecord{}, false, nil
	}

	// #### VARIANT 4: <GUESS>

	{
		ctx.PrintVerbose("Search for record (guess query type)")

		for _, rec := range records {
			v, err := models.UnmarshalPassword(ctx, rec)
			if err != nil {
				continue
			}

			if v.ID == query {
				return rec, v, true, nil
			}
			if v.Hostname == query {
				return rec, v, true, nil
			}
			if parsedURI != nil {
				if recordURI, err := url.Parse(v.Hostname); err == nil {
					if strings.ToLower(recordURI.Host) == strings.ToLower(parsedURI.Host) {
						return rec, v, true, nil
					}
				}
			}
		}
		return models.Record{}, models.PasswordRecord{}, false, nil
	}

}

func (a *CLIArgumentsPasswordsUtil) extUrlParse(v string) (*url.URL, error) {
	if !urlSchemaRegex.MatchString(v) {
		v = "generic://" + v
	}

	return url.Parse(v)
}

func (a *CLIArgumentsPasswordsList) filterDeleted(ctx *cli.FFSContext, records []models.PasswordRecord, includeDeleted bool, onlyDeleted bool) []models.PasswordRecord {
	result := make([]models.PasswordRecord, 0, len(records))

	for _, v := range records {
		if v.Deleted && !includeDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is deleted and include-deleted == false)", v.ID))
			continue
		}

		if !v.Deleted && onlyDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is not deleted and only-deleted == true)", v.ID))
			continue
		}

		result = append(result, v)
	}

	return result
}
