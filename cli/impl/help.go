package impl

import "C"
import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
	"strings"
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

func (a *CLIArgumentsHelp) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient <mode> --help", "Output specific help for a single action/verb"},
	}
}

func (a *CLIArgumentsHelp) FullHelp() []string {
	return []string{
		"$> ffsclient --help",
		"",
		"Show this help output.",
		"",
		"Can also be used as `ffsclient <mode> --help`",
	}
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

		leftlen := 0
		verbhelp := make([][]string, 0, 128)
		for _, mode := range cli.Modes {
			verb := GetModeImpl(mode)
			for _, line := range verb.ShortHelp() {
				left := line[0]
				right := line[1]
				if !strings.HasPrefix(left, "ffsclient") && strings.HasPrefix(left, "  ") && right != "" {
					right = "  # " + right
				}
				verbhelp = append(verbhelp, []string{left, right})
				leftlen = langext.Max(leftlen, len(left))
			}
		}

		ctx.PrintPrimaryOutput("firefox-sync-client.")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Usage:")
		for _, row := range verbhelp {
			ctx.PrintPrimaryOutput("  " + langext.StrPadRight(row[0], " ", leftlen) + "  " + row[1])
		}
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
		ctx.PrintPrimaryOutput("  --force-refresh-session                     Always auto-refresh the session, even if its not expired")
		return a.ExitCode

	} else {

		verb := GetModeImpl(*a.Verb)

		ctx.PrintPrimaryOutput("")
		for _, line := range verb.FullHelp() {
			ctx.PrintPrimaryOutput(line)
		}
		ctx.PrintPrimaryOutput("")

		return a.ExitCode

	}

}
