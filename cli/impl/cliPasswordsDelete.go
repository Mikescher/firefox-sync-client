package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsPasswordsDelete struct {
	Query            string
	QueryIsHost      bool
	QueryIsExactHost bool
	QueryIsID        bool
	HardDelete       bool

	CLIArgumentsPasswordsUtil
}

func NewCLIArgumentsPasswordsDelete() *CLIArgumentsPasswordsDelete {
	return &CLIArgumentsPasswordsDelete{
		Query:            "",
		QueryIsHost:      false,
		QueryIsExactHost: false,
		QueryIsID:        false,
		HardDelete:       false,
	}
}

func (a *CLIArgumentsPasswordsDelete) Mode() cli.Mode {
	return cli.ModePasswordsDelete
}

func (a *CLIArgumentsPasswordsDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsPasswordsDelete) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
}

func (a *CLIArgumentsPasswordsDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient passwords delete <host|id> [--hard]", "Delete a single password"},
		{"          [--is-host | --is-exact-host | --is-id]", "Specify that the supplied argument is a host / record-id (otherwise both is possible)"},
	}
}

func (a *CLIArgumentsPasswordsDelete) FullHelp() []string {
	return []string{
		"$> ffsclient passwords delete <host|id> [--is-host | --is-exact-host | --is-id] [--hard]",
		"",
		"Delete a single password",
		"",
		"By default we can supply a host or a record-id.",
		"If --is-host is specified, the query is parsed as an URI and we return the password that matches the host.",
		"If --is-exact-host is specified, the query is matched exactly against the host field in the password record.",
		"If --is-id is specified, the query is matched exactly agains the record-id.",
		"If --is-id is _not_ specified this method needs to query all passwords from the server and do a local search.",
		"If multiple passwords match the first match is deleted.",
		"If --hard is specified we delete the record, otherwise we only add {deleted:true} to mark it as a tombstone",
		"If no matching password is found the exitcode [82] is returned",
	}
}

func (a *CLIArgumentsPasswordsDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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
		if arg.Key == "hard" && arg.Value == nil {
			a.HardDelete = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsPasswordsDelete) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete Password]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("Query", a.Query)
	ctx.PrintVerboseKV("HardDelete", a.HardDelete)

	if langext.BoolCount(a.QueryIsID, a.QueryIsExactHost, a.QueryIsHost) > 1 {
		return fferr.NewDirectOutput(consts.ExitcodeError, "Must specify at most one of --id, --exact-host, --host")
	}
	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	_, record, found, err := a.findPasswordRecord(ctx, client, session, a.Query, a.QueryIsID, a.QueryIsHost, a.QueryIsExactHost)
	if err != nil {
		return err
	}

	if !found {
		return fferr.NewDirectOutput(consts.ExitcodePasswordNotFound, "Record not found")
	}

	ctx.PrintVerbose("Delete Record " + record.ID)

	if a.HardDelete {

		err = client.DeleteRecord(ctx, session, consts.CollectionPasswords, record.ID)
		if err != nil {
			return err
		}

	} else {

		err = client.SoftDeleteRecord(ctx, session, consts.CollectionPasswords, record.ID)
		if err != nil {
			return err
		}

	}

	// ========================================================================

	return a.printOutput(ctx, record)
}

func (a *CLIArgumentsPasswordsDelete) printOutput(ctx *cli.FFSContext, password models.PasswordRecord) error {
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
