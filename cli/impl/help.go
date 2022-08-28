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
		ctx.PrintPrimaryOutput("  ffsclient login <login> <password>          Login to FF-Sync account, uses ~/.config as default session location")
		ctx.PrintPrimaryOutput("                  [--device-name=<name>]")
		ctx.PrintPrimaryOutput("                  [--device-type=<type>]")
		ctx.PrintPrimaryOutput("  ffsclient refresh [--force]                 Refresh the current session (BID Assertion)")
		ctx.PrintPrimaryOutput("  ffsclient collections                       List all available collections")
		ctx.PrintPrimaryOutput("                  [--usage]                     # Include usage (storage space)")
		ctx.PrintPrimaryOutput("  ffsclient quota                             Query the storage quota of the current user")
		ctx.PrintPrimaryOutput("  ffsclient list <collection>                 Get a all records in a collection (use --format to define the format)")
		ctx.PrintPrimaryOutput("                  (--raw | --decoded | --ids) Return raw data, decoded payload, or only IDs")
		ctx.PrintPrimaryOutput("                  [--after <rfc3339>]         Return only fields updated after this date")
		ctx.PrintPrimaryOutput("                  [--sort <sort>]             Sort the result by (newest|index|oldest)")
		ctx.PrintPrimaryOutput("                  [--limit <n>]               Return max <n> elements")
		ctx.PrintPrimaryOutput("                  [--offset <o>]              Skip the first <n> elements")
		ctx.PrintPrimaryOutput("  ffsclient get <collection> <record-id>      Get a single record")
		ctx.PrintPrimaryOutput("                  (--raw | --decoded)         Return raw data or decoded payload")
		ctx.PrintPrimaryOutput("  ffsclient delete <record-id>                Delete the specified record")
		ctx.PrintPrimaryOutput("  ffsclient delete-all                        Delete all (!) records in the server")
		ctx.PrintPrimaryOutput("  ffsclient <verb> --help                     Output specific help for a single action/verb")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Options:")
		ctx.PrintPrimaryOutput("  -h, --help                                  Show this screen.")
		ctx.PrintPrimaryOutput("  -version                                    Show version.")
		ctx.PrintPrimaryOutput("  -v, --verbose                               Output more intermediate information")
		ctx.PrintPrimaryOutput("  -q, --quiet                                 Do not print anything")
		ctx.PrintPrimaryOutput("  --sessionfile <cfg>, --sessionfile=<cfg>    Specify the location of the saved session")
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
		ctx.PrintPrimaryOutput("  --no-autosave-session                       Do not update the sessionfile if the session was auto-refreshed")
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

		case cli.ModeLogin:
			ctx.PrintPrimaryOutput("ffsclient login <email> <password> [--device-name] [--device-type]")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Login to FF-Sync account")
			ctx.PrintPrimaryOutput("If no sesionfile location is provided this uses the default ~/.config/firefox-sync-client.secret")
			ctx.PrintPrimaryOutput("Specify a Device-name to identify the client in the Firefox Account page")

		case cli.ModeListCollections:
			ctx.PrintPrimaryOutput("ffsclient collections [--usage]")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("List all available collections together with their last-modified-time and entry-count")
			ctx.PrintPrimaryOutput("Optionally includes the storage-space usage (Note: This request may be very expensive)")
			return a.ExitCode

		case cli.ModeGetQuota:
			ctx.PrintPrimaryOutput("ffsclient quota")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Get the storage quota of the current user (used / max)")
			return a.ExitCode

		case cli.ModeTokenRefresh:
			ctx.PrintPrimaryOutput("ffsclient refresh [--force]")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Refresh the current session token")
			ctx.PrintPrimaryOutput("Use --force to force a new session, even if the old is still valid")
			return a.ExitCode

		case cli.ModeListRecords:
			ctx.PrintPrimaryOutput("  ffsclient list <collection> (--raw | --decoded | --ids) [--after <rfc3339>] [--sort <newest|index|oldest>]")
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("List all records in a collection")
			ctx.PrintPrimaryOutput("Either --raw or --decoded or --ids must be specified")
			ctx.PrintPrimaryOutput("If --after is specified (as an RFC 3339 timestamp) only records with an newer update-time are returned")
			ctx.PrintPrimaryOutput("If --sort is specified the resulting records are sorted by ( newest | index | oldest )")
			ctx.PrintPrimaryOutput("The global --format option is used to control the output format")
			return a.ExitCode

		case cli.ModeDeleteAll: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeDeleteSingle: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeGetRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeCreateRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		case cli.ModeUpdateRecord: //TODO
			ctx.PrintPrimaryOutput("")
			return a.ExitCode

		}

		panic("Unknnown verb: " + a.Verb.String())

	}

}
