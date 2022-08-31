package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
)

type CLIArgumentsBookmarksBase struct {
}

func NewCLIArgumentsBookmarksBase() *CLIArgumentsBookmarksBase {
	return &CLIArgumentsBookmarksBase{}
}

func (a *CLIArgumentsBookmarksBase) Mode() cli.Mode {
	return cli.ModeBookmarksBase
}

func (a *CLIArgumentsBookmarksBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsBookmarksBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsBookmarksBase) FullHelp() []string {
	r := []string{
		"$> ffsclient bookmarks (list|delete|create|update)",
		"======================================================",
		"",
	}
	for _, v := range ListSubcommands(a.Mode()) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsBookmarksBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient bookmarks must be called with a subcommand (eg `ffsclient bookmarks list`)")
}

func (a *CLIArgumentsBookmarksBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}
