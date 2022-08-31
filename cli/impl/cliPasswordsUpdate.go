package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
)

type CLIArgumentsPasswordsUpdate struct {
	Query            string
	QueryIsHost      bool
	QueryIsExactHost bool
	QueryIsID        bool

	NewHost          *string
	NewUsername      *string
	NewPassword      *string
	NewFormSubmitURL *string
	NewHTTPRealm     *string
	NewUsernameField *string
	NewPasswordField *string

	CLIArgumentsPasswordsUtil
}

func NewCLIArgumentsPasswordsUpdate() *CLIArgumentsPasswordsUpdate {
	return &CLIArgumentsPasswordsUpdate{
		Query:            "",
		QueryIsHost:      false,
		QueryIsExactHost: false,
		QueryIsID:        false,
		NewHost:          nil,
		NewUsername:      nil,
		NewPassword:      nil,
		NewFormSubmitURL: nil,
		NewHTTPRealm:     nil,
		NewUsernameField: nil,
		NewPasswordField: nil,
	}
}

func (a *CLIArgumentsPasswordsUpdate) Mode() cli.Mode {
	return cli.ModePasswordsUpdate
}

func (a *CLIArgumentsPasswordsUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsPasswordsUpdate) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient passwords update <host|id>", "Update an existing password"},
		{"          [--is-host | --is-exact-host | --is-id]", "Specify that the supplied argument is a host / record-id (otherwise both is possible)"},
		{"          [--host <url>]", "Update the host field"},
		{"          [--username <user>]", "Update the username"},
		{"          [--password <pass>]", "Update the password"},
		{"          [--form-submit-url <url>]", "Update the submission URL (GET/POST url set by <form>)"},
		{"          [--http-realm <realm>]", "Update the HTTP Realm (HTTP Realm for which the login is valid)"},
		{"          [--username-field <name>]", "Update the Username field (HTML field name of the username)"},
		{"          [--password-field <name>]", "Update the Password field (HTML field name of the password)"},
	}
}

func (a *CLIArgumentsPasswordsUpdate) FullHelp() []string {
	return []string{
		"$> ffsclient passwords update <host|id> [--is-host | --is-exact-host | --is-id] [--host <url>] [--username <user>] [--password <pass>] [--form-submit-url <url>] [--http-realm <realm>] [--username-field <name>] [--password-field <name>]",
		"",
		"Update an existing password",
		"",
		"By default we can supply a host or a record-id.",
		"If --is-host is specified, the query is parsed as an URI and we return the password that matches the host.",
		"If --is-exact-host is specified, the query is matched exactly against the host field in the password record.",
		"If --is-id is specified, the query is matched exactly agains the record-id.",
		"If --is-id is _not_ specified this method needs to query all passwords from the server and do a local search.",
		"If no matching password is found the exitcode [82] is returned",
		"",
		"The fields of the found password can be updated individually with the parameter:",
		"  * --host",
		"  * --username",
		"  * --password",
		"  * --form-submit-url",
		"  * --http-realm",
		"  * --username-field",
		"  * --password-field",
	}
}

func (a *CLIArgumentsPasswordsUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Query = positionalArgs[0]

	for _, arg := range optionArgs {
		if arg.Key == "is-host" && arg.Value == nil {
			a.QueryIsHost = true
			continue
		}
		if arg.Key == "is-exact-host" && arg.Value == nil {
			a.QueryIsExactHost = true
			continue
		}
		if arg.Key == "is-id" && arg.Value == nil {
			a.QueryIsID = true
			continue
		}
		if arg.Key == "host" && arg.Value != nil {
			a.NewHost = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "username" && arg.Value != nil {
			a.NewUsername = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "password" && arg.Value != nil {
			a.NewPassword = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "form-submit-url" && arg.Value != nil {
			a.NewFormSubmitURL = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "http-realm" && arg.Value != nil {
			a.NewHTTPRealm = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "username-field" && arg.Value != nil {
			a.NewUsernameField = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "password-field" && arg.Value != nil {
			a.NewPasswordField = langext.Ptr(*arg.Value)
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsUpdate) Execute(ctx *cli.FFSContext) int {
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

	record, found, err := a.findPasswordRecord(ctx, client, session, a.Query, a.QueryIsID, a.QueryIsHost, a.QueryIsExactHost)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !found {
		ctx.PrintErrorMessage("Record not found")
		return consts.ExitcodePasswordNotFound
	}

	ctx.PrintVerbose("Update Record " + record.ID)

	record = a.updateRecordFields(ctx, record)

	plain, err := record.ToPlaintextPayload()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	payload, err := client.EncryptPayload(ctx, session, consts.CollectionPasswords, plain)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	err = client.PutRecord(ctx, session, consts.CollectionPasswords, record.ID, payload, false, false)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	return a.printOutput(ctx, record)
}

func (a *CLIArgumentsPasswordsUpdate) printOutput(ctx *cli.FFSContext, password models.PasswordRecord) int {
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

func (a *CLIArgumentsPasswordsUpdate) updateRecordFields(ctx *cli.FFSContext, record models.PasswordRecord) models.PasswordRecord {

	if a.NewHost != nil {
		ctx.PrintVerbose(fmt.Sprintf("Update Host from '%s' to '%s'", record.Hostname, *a.NewHost))
		record.Hostname = *a.NewHost
	}
	if a.NewUsername != nil {
		ctx.PrintVerbose(fmt.Sprintf("Update Username from '%s' to '%s'", record.Username, *a.NewUsername))
		record.Username = *a.NewUsername
	}
	if a.NewPassword != nil {
		ctx.PrintVerbose(fmt.Sprintf("Update Password from '%s' to '%s'", record.Password, *a.NewPassword))
		record.Password = *a.NewPassword
	}
	if a.NewFormSubmitURL != nil {
		ctx.PrintVerbose(fmt.Sprintf("Update FormSubmitURL from '%s' to '%s'", record.FormSubmitURL, *a.NewFormSubmitURL))
		record.FormSubmitURL = *a.NewFormSubmitURL
	}
	if a.NewHTTPRealm != nil {
		newrealm := langext.Ptr(*a.NewHTTPRealm)
		if *newrealm == "" {
			newrealm = nil
		}

		ctx.PrintVerbose(fmt.Sprintf("Update HTTPRealm from '%v' to '%v'", record.HTTPRealm, newrealm))
		record.HTTPRealm = newrealm
	}
	if a.NewUsernameField != nil {
		ctx.PrintVerbose(fmt.Sprintf("Update UsernameField from '%s' to '%s'", record.UsernameField, *a.NewUsernameField))
		record.UsernameField = *a.NewUsernameField
	}
	if a.NewPasswordField != nil {
		ctx.PrintVerbose(fmt.Sprintf("Update PasswordField from '%s' to '%s'", record.PasswordField, *a.NewPasswordField))
		record.PasswordField = *a.NewPasswordField
	}

	return record
}
