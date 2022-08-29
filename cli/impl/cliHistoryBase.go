package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"github.com/joomcode/errorx"
)

type CLIArgumentsHistoryBase struct {
}

func NewCLIArgumentsHistoryBase() *CLIArgumentsHistoryBase {
	return &CLIArgumentsHistoryBase{}
}

func (a *CLIArgumentsHistoryBase) Mode() cli.Mode {
	return cli.ModeHistoryBase
}

func (a *CLIArgumentsHistoryBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsHistoryBase) FullHelp() []string {
	r := []string{
		"$> ffsclient history (list|delete|create|update)",
		"======================================================",
		"",
	}
	for _, v := range ListSubcommands(a.Mode()) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsHistoryBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return errorx.InternalError.New("ffsclient history must be called with a subcommand (eg `ffsclient history list`)")
}

func (a *CLIArgumentsHistoryBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}
