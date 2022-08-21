package impl

import "C"
import (
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
)

type CLIArgumentsHelp struct {
	Extra    string
	Verb     *cli.Mode
	ExitCode int
}

func NewCLIArgumentsHelp() *CLIArgumentsHelp {
	return &CLIArgumentsHelp{
		Extra:    "",
		Verb:     nil,
		ExitCode: 0,
	}
}

func (a *CLIArgumentsHelp) Mode() cli.Mode {
	return cli.ModeHelp
}

func (a *CLIArgumentsHelp) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	if len(optionArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + optionArgs[0].Key)
	}

	return nil
}

func (a *CLIArgumentsHelp) Execute(ctx *cli.FFSContext) int {
	if a.Extra != "" {
		ctx.PrintPrimaryOutput(a.Extra)
		ctx.PrintPrimaryOutput("")
	}

	// http://docopt.org/

	if a.Verb == nil {

		ctx.PrintPrimaryOutput("firefox-sync-client.")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Usage:")
		ctx.PrintPrimaryOutput("  ffsclient login <login> <password>          Login to FF-Sync account, uses ~/.config as default config location")
		ctx.PrintPrimaryOutput("                  [--sevice-name=<name>]")
		ctx.PrintPrimaryOutput("                  [--sevice-name <name>]")
		ctx.PrintPrimaryOutput("  ffsclient delete-all                        Delete all (!) records in the server")
		ctx.PrintPrimaryOutput("  ffsclient delete <record-id>                Delete the specified record")
		ctx.PrintPrimaryOutput("  ffsclient collections                       List all available collections")
		ctx.PrintPrimaryOutput("  ffsclient raw <collection> <record-id>      get a single record (not decoded)")
		ctx.PrintPrimaryOutput("  ffsclient get <collection> <record-id>      get a single record (decoded)")
		ctx.PrintPrimaryOutput("  ffsclient create <collection> ...           (TODO)")
		ctx.PrintPrimaryOutput("  ffsclient update <collection> ...           (TODO)")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Options:")
		ctx.PrintPrimaryOutput("  -h, --help                                  Show this screen.")
		ctx.PrintPrimaryOutput("  -version                                    Show version.")
		ctx.PrintPrimaryOutput("  -v, --verbose                               Output more intermediate information")
		ctx.PrintPrimaryOutput("  -q, --quiet                                 Do not print anything")
		ctx.PrintPrimaryOutput("  -c <f>, --config <conf>, --config=<conf>    Specify the config location")
		ctx.PrintPrimaryOutput("  -f <fmt>, --format <fmt>, --format=<fmt>    Specify the output format")
		ctx.PrintPrimaryOutput("                                                # - 'text'")
		ctx.PrintPrimaryOutput("                                                # - 'json'")
		ctx.PrintPrimaryOutput("                                                # - 'netscape'     (default firefox bookmarks format)")
		ctx.PrintPrimaryOutput("                                                # - 'bookmarksxml' (custom XML bookmarks format)")
		ctx.PrintPrimaryOutput("  --auth-server <url>, --auth-server=<url>    Specify the (authentication) server-url")
		ctx.PrintPrimaryOutput("  --token-server <url>, --token-server=<url>  Specify the (token) server-url")
		ctx.PrintPrimaryOutput("  --color                                     Enforce colored output")
		ctx.PrintPrimaryOutput("  --no-color                                  Disable colored output")
		return a.ExitCode

	} else {

		switch *a.Verb {

		case cli.ModeHelp:
			ctx.PrintPrimaryOutput("ffsclient help")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Show this help output.")
			ctx.PrintPrimaryOutput("Can also be used as `ffsclient <verb> --help`")
			return a.ExitCode

		case cli.ModeVersion:
			ctx.PrintPrimaryOutput("ffsclient version")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Login to FF-Sync account")
			ctx.PrintPrimaryOutput("If no config location is provided this uses the default ~/.config/firefox-sync-client.secret")

		case cli.ModeLogin:
			ctx.PrintPrimaryOutput("ffsclient login <email> <password>")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Login to FF-Sync account")
			ctx.PrintPrimaryOutput("If no config location is provided this uses the default ~/.config/firefox-sync-client.secret")

		case cli.ModeDeleteAll:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeDeleteSingle:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeListCollections:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeGetRawRecord:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeGetDecodedRecord:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeCreateRecord:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeUpdateRecord:
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		}

		panic("Unknnown verb: " + a.Verb.String())

	}

}
