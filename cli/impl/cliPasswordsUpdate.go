package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
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

func (a *CLIArgumentsPasswordsUpdate) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
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
		"Update the specified fields of an existing password entry.",
		"",
		"By default we can supply a host or a record-id.",
		"If --is-host is specified, the query is parsed as an URI and we return the password that matches the host.",
		"If --is-exact-host is specified, the query is matched exactly against the host field in the password record.",
		"If --is-id is specified, the query is matched exactly agains the record-id.",
		"If --is-id is _not_ specified this method needs to query all passwords from the server and do a local search.",
		"If no matching password is found the exitcode [82] is returned",
		"",
		"The fields of the found password can be updated individually with the parameters:",
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

func (a *CLIArgumentsPasswordsUpdate) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete Password]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Query", a.Query)

	if langext.BoolCount(a.QueryIsID, a.QueryIsExactHost, a.QueryIsHost) > 1 {
		return fferr.NewDirectOutput(consts.ExitcodeError, "Must specify at most one of --id, --exact-host, --host")
	}
	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[0] Find Record")

	record, pwrec, found, err := a.findPasswordRecord(ctx, client, session, a.Query, a.QueryIsID, a.QueryIsHost, a.QueryIsExactHost)
	if err != nil {
		return err
	}

	if !found {
		return fferr.NewDirectOutput(consts.ExitcodePasswordNotFound, "Record not found")
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[2] Patch Data")

	newData, err := a.patchData(ctx, record, pwrec)
	if err != nil {
		return err
	}

	// ========================================================================

	if string(newData) != string(record.DecodedData) {

		ctx.PrintVerboseHeader("[3] Update record")

		newPayloadRecord, err := client.EncryptPayload(ctx, session, consts.CollectionPasswords, string(newData))
		if err != nil {
			return err
		}

		update := models.RecordUpdate{
			ID:      record.ID,
			Payload: langext.Ptr(newPayloadRecord),
		}

		err = client.PutRecord(ctx, session, consts.CollectionPasswords, update, false, false)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.WrapDirectOutput(err, consts.ExitcodeRecordNotFound, "Record not found")
		}
		if err != nil {
			return err
		}

	} else {

		ctx.PrintVerbose("Do not update record (nothing to do)")

	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput("Okay.")
	return nil
}

func (a *CLIArgumentsPasswordsUpdate) printOutput(ctx *cli.FFSContext, password models.PasswordRecord) error {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(password.ID)
		return nil

	case cli.OutputFormatJson:
		ctx.PrintPrimaryOutputJSON(password.ToJSON(ctx, true))
		return nil

	case cli.OutputFormatXML:
		ctx.PrintPrimaryOutputXML(password.ToXML(ctx, "Password", true))
		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}

func (a *CLIArgumentsPasswordsUpdate) patchData(ctx *cli.FFSContext, record models.Record, pwrec models.PasswordRecord) ([]byte, error) {
	var err error

	newData := record.DecodedData

	if a.NewHost != nil {
		ctx.PrintVerbose(fmt.Sprintf("Patch field [hostname] from \"%s\" to \"%s\"", pwrec.Hostname, *a.NewHost))

		newData, err = langext.PatchJson(newData, "hostname", *a.NewHost)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.NewUsername != nil {
		ctx.PrintVerbose(fmt.Sprintf("Patch field [username] from \"%s\" to \"%s\"", pwrec.Username, *a.NewUsername))

		newData, err = langext.PatchJson(newData, "username", *a.NewUsername)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.NewPassword != nil {
		ctx.PrintVerbose(fmt.Sprintf("Patch field [password] from \"%s\" to \"%s\"", pwrec.Password, *a.NewPassword))

		newData, err = langext.PatchJson(newData, "password", *a.NewPassword)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.NewFormSubmitURL != nil {
		ctx.PrintVerbose(fmt.Sprintf("Patch field [formSubmitURL] from \"%s\" to \"%s\"", pwrec.FormSubmitURL, *a.NewFormSubmitURL))

		newData, err = langext.PatchJson(newData, "formSubmitURL", *a.NewFormSubmitURL)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.NewHTTPRealm != nil {
		if *a.NewHTTPRealm != "" {
			ctx.PrintVerbose(fmt.Sprintf("Patch field [httpRealm] from \"%v\" to \"%v\"", pwrec.HTTPRealm, *a.NewHTTPRealm))

			newData, err = langext.PatchJson(newData, "httpRealm", *a.NewHTTPRealm)
			if err != nil {
				return nil, errorx.Decorate(err, "failed to patch data of existing record")
			}
		} else {
			ctx.PrintVerbose(fmt.Sprintf("Patch field [httpRealm] from \"%v\" to \"%v\"", pwrec.HTTPRealm, a.NewHTTPRealm))

			newData, err = langext.PatchRemJson(newData, "httpRealm")
			if err != nil {
				return nil, errorx.Decorate(err, "failed to patch data of existing record")
			}
		}
	}

	if a.NewUsernameField != nil {
		ctx.PrintVerbose(fmt.Sprintf("Patch field [usernameField] from \"%s\" to \"%s\"", pwrec.UsernameField, *a.NewUsernameField))

		newData, err = langext.PatchJson(newData, "usernameField", *a.NewUsernameField)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.NewPasswordField != nil {
		ctx.PrintVerbose(fmt.Sprintf("Patch field [passwordField] from \"%s\" to \"%s\"", pwrec.PasswordField, *a.NewPasswordField))

		newData, err = langext.PatchJson(newData, "passwordField", *a.NewPasswordField)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	return newData, nil
}
