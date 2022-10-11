package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"fmt"
)

type CLIArgumentsFormsBase struct {
}

func NewCLIArgumentsFormsBase() *CLIArgumentsFormsBase {
	return &CLIArgumentsFormsBase{}
}

func (a *CLIArgumentsFormsBase) Mode() cli.Mode {
	return cli.ModeFormsBase
}

func (a *CLIArgumentsFormsBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsFormsBase) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsFormsBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsFormsBase) FullHelp() []string {
	r := []string{
		"$> ffsclient forms (list|delete|create|update|get)",
		"======================================================",
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

func (a *CLIArgumentsFormsBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient forms must be called with a subcommand (eg `ffsclient forms list`)")
}

func (a *CLIArgumentsFormsBase) Execute(ctx *cli.FFSContext) error {
	return fferr.NewDirectOutput(consts.ExitcodeError, "Cannot call `forms` command without an subcommand")
}

type CLIArgumentsFormsUtil struct{}

func (a *CLIArgumentsFormsUtil) filterDeleted(ctx *cli.FFSContext, records []models.FormRecord, includeDeleted bool, onlyDeleted bool, name *[]string) []models.FormRecord {
	result := make([]models.FormRecord, 0, len(records))

	for _, v := range records {
		if v.Deleted && !includeDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is deleted and include-deleted == false)", v.ID))
			continue
		}

		if !v.Deleted && onlyDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is not deleted and only-deleted == true)", v.ID))
			continue
		}

		if name != nil && !langext.InArray(v.Name, *name) {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (not in name-filter)", v.ID))
			continue
		}

		result = append(result, v)
	}

	return result
}

func (a *CLIArgumentsFormsUtil) newFormID() string {
	return langext.RandBase62(16)
}
