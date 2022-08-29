package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"github.com/joomcode/errorx"
	"time"
)

type CLIArgumentsPasswordsBase struct {
	ShowPasswords      bool
	IgnoreSchemaErrors bool
	Sort               *string
	Limit              *int
	Offset             *int
	After              *time.Time
}

func NewCLIArgumentsPasswordsBase() *CLIArgumentsPasswordsBase {
	return &CLIArgumentsPasswordsBase{
		ShowPasswords: false,
		Sort:          nil,
		Limit:         nil,
		Offset:        nil,
		After:         nil,
	}
}

func (a *CLIArgumentsPasswordsBase) Mode() cli.Mode {
	return cli.ModePasswordsBase
}

func (a *CLIArgumentsPasswordsBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsPasswordsBase) FullHelp() []string {
	r := []string{
		"$> ffsclient passwords (list|delete|create|update|get)",
		"======================================================",
		"",
	}
	for _, v := range ListSubcommands(a.Mode()) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsPasswordsBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return errorx.InternalError.New("ffsclient passwords must be called with a subcommand (eg `ffsclient passwords list`)")
}

func (a *CLIArgumentsPasswordsBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}
