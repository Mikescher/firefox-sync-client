package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"fmt"
)

type CLIArgumentsHistoryBase struct {
	CLIArgumentsHistoryUtil
}

func NewCLIArgumentsHistoryBase() *CLIArgumentsHistoryBase {
	return &CLIArgumentsHistoryBase{}
}

func (a *CLIArgumentsHistoryBase) Mode() cli.Mode {
	return cli.ModeHistoryBase
}

func (a *CLIArgumentsHistoryBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsHistoryBase) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsHistoryBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsHistoryBase) FullHelp() []string {
	r := []string{
		"$> ffsclient history (list|delete|create|update)",
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

func (a *CLIArgumentsHistoryBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient history must be called with a subcommand (eg `ffsclient history list`)")
}

func (a *CLIArgumentsHistoryBase) Execute(ctx *cli.FFSContext) error {
	return fferr.NewDirectOutput(consts.ExitcodeError, "Cannot call `history` command without an subcommand")
}

type CLIArgumentsHistoryUtil struct {
	CLIArgumentsBaseUtil
}

func (a *CLIArgumentsHistoryUtil) filterDeleted(ctx *cli.FFSContext, records []models.HistoryRecord, includeDeleted bool, onlyDeleted bool) []models.HistoryRecord {
	result := make([]models.HistoryRecord, 0, len(records))

	for _, v := range records {
		if v.Deleted && !includeDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is deleted and include-deleted == false)", v.ID))
			continue
		}

		if !v.Deleted && onlyDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is not deleted and only-deleted == true)", v.ID))
			continue
		}

		result = append(result, v)
	}

	return result
}

func (a *CLIArgumentsHistoryUtil) newHistoryID() string {
	return langext.RandBase62(12)
}
