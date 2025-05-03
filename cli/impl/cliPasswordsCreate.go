package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"github.com/joomcode/errorx"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"time"
)

type CLIArgumentsPasswordsCreate struct {
	Host          string
	Username      string
	Password      string
	FormSubmitURL *string
	HTTPRealm     *string
	UsernameField *string
	PasswordField *string

	CLIArgumentsPasswordsUtil
}

func NewCLIArgumentsPasswordsCreate() *CLIArgumentsPasswordsCreate {
	return &CLIArgumentsPasswordsCreate{}
}

func (a *CLIArgumentsPasswordsCreate) Mode() cli.Mode {
	return cli.ModePasswordsCreate
}

func (a *CLIArgumentsPasswordsCreate) PositionArgCount() (*int, *int) {
	return langext.Ptr(3), langext.Ptr(3)
}

func (a *CLIArgumentsPasswordsCreate) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsPasswordsCreate) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient passwords create <host> <username> <password>", "Insert a new password"},
		{"          [--form-submit-url <url>]", "Specify the submission URL (GET/POST url set by <form>)"},
		{"          [--http-realm <realm>]", "Specify the HTTP Realm (HTTP Realm for which the login is valid)"},
		{"          [--username-field <name>]", "Specify the Username field (HTML field name of the username)"},
		{"          [--password-field <name>]", "Specify the Password field (HTML field name of the password)"},
	}
}

func (a *CLIArgumentsPasswordsCreate) FullHelp() []string {
	return []string{
		"$> ffsclient passwords create <host> <username> <password> [--form-submit-url <url>] [--http-realm <realm>] [--username-field <name>] [--password-field <name>]",
		"",
		"Insert a new password",
		"",
		"The fields <host>, <username> <password> must be specified.",
		"The fields formSubmitURL, httpRealm, usernameField and passwordField are optional and get their default values (empty-string/null) if not supplied",
		"",
		"Outputs the RecordID of the newly created entry on success.",
	}
}

func (a *CLIArgumentsPasswordsCreate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Host = positionalArgs[0]
	a.Username = positionalArgs[1]
	a.Password = positionalArgs[2]

	for _, arg := range optionArgs {
		if arg.Key == "form-submit-url" && arg.Value != nil {
			a.FormSubmitURL = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "http-realm" && arg.Value != nil {
			a.HTTPRealm = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "username-field" && arg.Value != nil {
			a.UsernameField = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "password-field" && arg.Value != nil {
			a.PasswordField = langext.Ptr(*arg.Value)
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsCreate) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Create Password]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	recordID := a.newPasswordID()

	now := time.Now()

	bso := models.PasswordPayloadSchema{
		ID:                  recordID,
		Hostname:            a.Host,
		FormSubmitURL:       langext.Coalesce(a.FormSubmitURL, ""),
		HTTPRealm:           a.HTTPRealm,
		Username:            a.Username,
		Password:            a.Password,
		UsernameField:       langext.Coalesce(a.UsernameField, ""),
		PasswordField:       langext.Coalesce(a.PasswordField, ""),
		TimeCreated:         langext.Ptr(now.UnixMilli()),
		TimePasswordChanged: langext.Ptr(now.UnixMilli()),
	}

	plainPayload, err := json.Marshal(bso)
	if err != nil {
		return errorx.Decorate(err, "failed to marshal BSO json")
	}

	payload, err := client.EncryptPayload(ctx, session, consts.CollectionPasswords, string(plainPayload))
	if err != nil {
		return err
	}

	update := models.RecordUpdate{
		ID:      recordID,
		Payload: langext.Ptr(payload),
	}

	err = client.PutRecord(ctx, session, consts.CollectionPasswords, update, false, false)
	if err != nil {
		return err
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput(recordID)
	return nil
}
