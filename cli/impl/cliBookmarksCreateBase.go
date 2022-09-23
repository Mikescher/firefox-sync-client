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

func (a *CLIArgumentsBookmarksCreateBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsBookmarksCreateBase) FullHelp() []string {
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

func (a *CLIArgumentsBookmarksCreateBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient bookmarks create must be called with a specific type (eg `ffsclient bookmarks create folder`)")
}

func (a *CLIArgumentsBookmarksCreateBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}
