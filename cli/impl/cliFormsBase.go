package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"github.com/joomcode/errorx"
)

type CLIArgumentsFormsBase struct {
}

func NewCLIArgumentsFormsBase() *CLIArgumentsFormsBase {
	return &CLIArgumentsFormsBase{}
}

func (a *CLIArgumentsFormsBase) Mode() cli.Mode {
	return cli.ModeFormsBase
}

func (a *CLIArgumentsFormsBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsFormsBase) FullHelp() []string {
	r := []string{
		"$> ffsclient forms (list|delete|create|update|get)",
		"======================================================",
		"",
	}
	for _, v := range ListSubcommands(a.Mode()) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsFormsBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return errorx.InternalError.New("ffsclient forms must be called with a subcommand (eg `ffsclient forms list`)")
}

func (a *CLIArgumentsFormsBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}
