package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/google/uuid"
	"github.com/joomcode/errorx"
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

func (a *CLIArgumentsPasswordsCreate) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Create Password]")
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

	recordID := "{" + uuid.New().String() + "}"

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
		ctx.PrintFatalError(errorx.Decorate(err, "failed to marshal BSO json"))
		return consts.ExitcodeError
	}

	payload, err := client.EncryptPayload(ctx, session, consts.CollectionPasswords, string(plainPayload))
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	update := models.RecordUpdate{
		ID:      recordID,
		Payload: langext.Ptr(payload),
	}

	err = client.PutRecord(ctx, session, consts.CollectionPasswords, update, false, false)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput(recordID)
	return 0
}
