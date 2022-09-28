package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
)

type CLIArgumentsPasswordsGet struct {
	Query            string
	QueryIsHost      bool
	QueryIsExactHost bool
	QueryIsID        bool

	CLIArgumentsPasswordsUtil
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
		{"          [--is-host | --is-exact-host | --is-id]", "Specify that the supplied argument is a host / record-id (otherwise both is possible)"},
	}
}

func (a *CLIArgumentsPasswordsGet) FullHelp() []string {
	return []string{
		"$> ffsclient passwords get <host|id> [--is-host | --is-exact-host | --is-id]",
		"",
		"Get a password",
		"",
		"By default we can supply a host or a record-id.",
		"If --is-host is specified, the query is parsed as an URI and we return the password that matches the host.",
		"If --is-exact-host is specified, the query is matched exactly against the host field in the password record.",
		"If --is-id is specified, the query is matched exactly agains the record-id.",
		"If --is-id is _not_ specified this method needs to query all passwords from the server and do a local search.",
		"If no matching password is found the exitcode [82] is returned",
	}
}

func (a *CLIArgumentsPasswordsGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
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

	_, record, found, err := a.findPasswordRecord(ctx, client, session, a.Query, a.QueryIsID, a.QueryIsHost, a.QueryIsExactHost)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !found {
		ctx.PrintErrorMessage("Record not found")
		return consts.ExitcodePasswordNotFound
	}

	// ========================================================================

	return a.printOutput(ctx, record)
}

func (a *CLIArgumentsPasswordsGet) printOutput(ctx *cli.FFSContext, password models.PasswordRecord) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		ctx.PrintPrimaryOutput(password.Password)
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
