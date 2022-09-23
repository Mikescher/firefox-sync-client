package impl

import "C"
import (
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
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

func (a *CLIArgumentsHelp) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsHelp) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient <sub> --help", "Output specific help for a single subcommand"},
	}
}

func (a *CLIArgumentsHelp) FullHelp() []string {
	return []string{
		"$> ffsclient --help",
		"",
		"Show this help output.",
		"",
		"Can also be used as `ffsclient <sub> --help`",
	}
}

func (a *CLIArgumentsHelp) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(optionArgs) > 0 {
		return fferr.DirectOutput.New("Unknown argument: " + optionArgs[0].Key)
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
		opthelp := make([][]string, 0, 128)

		for _, mode := range cli.Modes {
			verb := GetModeImpl(mode)
			for _, line := range verb.ShortHelp() {
				left := line[0]
				right := line[1]
				if !strings.HasPrefix(left, "ffsclient") && (strings.HasPrefix(left, "  ") || left == "") && right != "" {
					right = "  # " + right
				}
				verbhelp = append(verbhelp, []string{left, right})
				leftlen = langext.Max(leftlen, len(left))
			}
		}

		for _, line := range a.globalOptions() {
			left := line[0]
			right := line[1]
			if !strings.HasPrefix(left, "-") && (strings.HasPrefix(left, "  ") || left == "") && right != "" {
				right = "  # " + right
			}
			opthelp = append(opthelp, []string{left, right})
			leftlen = langext.Max(leftlen, len(left))
		}

		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("firefox-sync-client.")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("# (Use `ffsclient <command> --help` for more detailed info)")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Usage:")
		for _, row := range verbhelp {
			ctx.PrintPrimaryOutput("  " + langext.StrPadRight(row[0], " ", leftlen) + "  " + row[1])
		}
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Hint:")
		ctx.PrintPrimaryOutput("  # If you need to supply a record-id / collection that starts with an minus, use the --!arg=... syntax")
		ctx.PrintPrimaryOutput("  #     e.g.: `ffsclient get bookmarks --!arg=-udhG86-JgpUx --decoded`")
		ctx.PrintPrimaryOutput("  # Also if you need to supply a argument that starts with an - use the --arg=value syntax")
		ctx.PrintPrimaryOutput("  #     e.g.: `ffsclient bookmarks add Test \"https://example.org\" --parent toolbar --position=-3`")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Options:")
		for _, row := range opthelp {
			ctx.PrintPrimaryOutput("  " + langext.StrPadRight(row[0], " ", leftlen) + "  " + row[1])
		}
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("Exit Codes:")
		ctx.PrintPrimaryOutput("  0             Program exited successfully")
		ctx.PrintPrimaryOutput("  60            Program existed with an (unspecified) error")
		ctx.PrintPrimaryOutput("  61            Program crashed")
		ctx.PrintPrimaryOutput("  62            Program called without arguments")
		ctx.PrintPrimaryOutput("  63            Failed to parse commandline arguments")
		ctx.PrintPrimaryOutput("  64            Command needs a valid session/session-file and none was found")
		ctx.PrintPrimaryOutput("  65            The current subcommand does not support the specified output format")
		ctx.PrintPrimaryOutput("  66            Record with this ID not found")
		ctx.PrintPrimaryOutput("")
		ctx.PrintPrimaryOutput("  81            (check-session): The session is not valid")
		ctx.PrintPrimaryOutput("  82            (passwords): No matching password found")
		ctx.PrintPrimaryOutput("  83            (create-bookmarks): Parent record is not a folder")
		ctx.PrintPrimaryOutput("  84            (create-bookmarks): The position in the parent would be out of bounds")
		ctx.PrintPrimaryOutput("  85            (update-bookmarks): One of the specified fields is not valid on the record type")
		ctx.PrintPrimaryOutput("")
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

func (a *CLIArgumentsHelp) globalOptions() [][]string { //TODO use yyyy-MM-dd syntax and convert
	return [][]string{
		{"-h, --help", "Show this screen."},
		{"--version", "Show version."},
		{"-v, --verbose", "Output more intermediate information"},
		{"-q, --quiet", "Do not print anything"},
		{"--sessionfile <cfg>, --sessionfile=<cfg>", "Specify the location of the saved session"},
		{"-f <fmt>, --format <fmt>, --format=<fmt>", "Specify the output format (not all subcommands support all output-formats)"},
		{"", "- 'text'"},
		{"", "- 'json'"},
		{"", "- 'netscape'   (default firefox bookmarks format)"},
		{"", "- 'xml'"},
		{"", "- 'table'"},
		{"--auth-server <url>, --auth-server=<url>", "Specify the (authentication) server-url"},
		{"--token-server <url>, --token-server=<url>", "Specify the (token) server-url"},
		{"--color", "Enforce colored output"},
		{"--no-color", "Disable colored output"},
		{"--timezone <tz>, --timezone=<tz>", "Specify the output timezone"},
		{"", "Can be either:"},
		{"", "  - UTC"},
		{"", "  - Local (default)"},
		{"", "  - IANA Time Zone, e.g. 'America/New_York'"},
		{"--timeformat <url>, --timeformat=<url>", "Specify the output timeformat (golang syntax)"},
		{"-o <f>, --output <f>, --output=<f>", "Write the output to a file"},
		{"--no-autosave-session", "Do not update the sessionfile if the session was auto-refreshed"},
		{"--force-refresh-session", "Always auto-refresh the session, even if its not expired"},
		{"--no-xml-declaration", "Do not print the xml declaration when using `--format xml`"},
	}
}
