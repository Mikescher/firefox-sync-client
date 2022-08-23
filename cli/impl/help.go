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
		ctx.PrintPrimaryOutput("                  [--service-name=<name>]")
		ctx.PrintPrimaryOutput("                  [--service-name <name>]")
		ctx.PrintPrimaryOutput("  ffsclient delete-all                        Delete all (!) records in the server")
		ctx.PrintPrimaryOutput("  ffsclient delete <record-id>                Delete the specified record")
		ctx.PrintPrimaryOutput("  ffsclient collections                       List all available collections")
		ctx.PrintPrimaryOutput("                  [--usage]                     # Include usage (storage space)")
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
		ctx.PrintPrimaryOutput("  -c <cfg>, --config <cfg>, --config=<cfg>    Specify the config location")
		ctx.PrintPrimaryOutput("  -f <fmt>, --format <fmt>, --format=<fmt>    Specify the output format")
		ctx.PrintPrimaryOutput("                                                # - 'text'")
		ctx.PrintPrimaryOutput("                                                # - 'json'")
		ctx.PrintPrimaryOutput("                                                # - 'netscape'   (default firefox bookmarks format)")
		ctx.PrintPrimaryOutput("                                                # - 'xml'")
		ctx.PrintPrimaryOutput("                                                # - 'table'")
		ctx.PrintPrimaryOutput("  --auth-server <url>, --auth-server=<url>    Specify the (authentication) server-url")
		ctx.PrintPrimaryOutput("  --token-server <url>, --token-server=<url>  Specify the (token) server-url")
		ctx.PrintPrimaryOutput("  --color                                     Enforce colored output")
		ctx.PrintPrimaryOutput("  --no-color                                  Disable colored output")
		ctx.PrintPrimaryOutput("  --timezone <url>, --timezone=<url>          Specify the output timezone")
		ctx.PrintPrimaryOutput("                                                # Can be either:")
		ctx.PrintPrimaryOutput("                                                #   - UTC")
		ctx.PrintPrimaryOutput("                                                #   - Local (default)")
		ctx.PrintPrimaryOutput("                                                #   - IANA Time Zone, e.g. 'America/New_York'")
		ctx.PrintPrimaryOutput("  --timeformat <url>, --timeformat=<url>      Specify the output timeformat (golang syntax)") //TODO use yyyy-MM-dd syntax and convert
		ctx.PrintPrimaryOutput("  -o <f>, --output <f>, --output=<f>          Write the output to a file")
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
			ctx.PrintPrimaryOutput("ffsclient login <email> <password> [--service-name]")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Login to FF-Sync account")
			ctx.PrintPrimaryOutput("If no config location is provided this uses the default ~/.config/firefox-sync-client.secret")
			ctx.PrintPrimaryOutput("Specify a service-name to identify the client in the Firefox Account page")

		case cli.ModeListCollections:
			ctx.PrintPrimaryOutput("ffsclient collections [--usage]")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("List all available collections together with their last-modified-time and entry-count")
			ctx.PrintPrimaryOutput("Optionally includes the storage-space usage (Note: This request may be very expensive)")
			return a.ExitCode

		case cli.ModeDeleteAll: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeDeleteSingle: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeGetRawRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeGetDecodedRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeCreateRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeUpdateRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeGetQuota: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		}

		panic("Unknnown verb: " + a.Verb.String())

	}

}
