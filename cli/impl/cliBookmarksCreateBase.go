package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
)

type CLIArgumentsBookmarksCreateBase struct {
	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateBase() *CLIArgumentsBookmarksCreateBase {
	return &CLIArgumentsBookmarksCreateBase{}
}

func (a *CLIArgumentsBookmarksCreateBase) Mode() cli.Mode {
	return cli.ModeBookmarksCreateBase
}

func (a *CLIArgumentsBookmarksCreateBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsBookmarksCreateBase) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsBookmarksCreateBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsBookmarksCreateBase) FullHelp() []string {
	r := []string{
		"$> ffsclient bookmarks create (folder|bookmark|separator)",
		"=========================================================",
		"",
		"",
	}
	for _, v := range ListSubcommands(a.Mode(), true) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
		r = append(r, "")
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsBookmarksCreateBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient bookmarks create must be called with a specific type (eg `ffsclient bookmarks create folder`), possible types are [bookmark | folder | separator]")
}

func (a *CLIArgumentsBookmarksCreateBase) Execute(ctx *cli.FFSContext) error {
	return fferr.NewDirectOutput(consts.ExitcodeError, "Cannot call `bookmarks` command without an subcommand")
}
