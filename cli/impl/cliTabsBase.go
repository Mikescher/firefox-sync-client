package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsTabsBase struct {
	CLIArgumentsTabsUtil
}

func NewCLIArgumentsTabsBase() *CLIArgumentsTabsBase {
	return &CLIArgumentsTabsBase{}
}

func (a *CLIArgumentsTabsBase) Mode() cli.Mode {
	return cli.ModeTabsBase
}

func (a *CLIArgumentsTabsBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsTabsBase) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsTabsBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsTabsBase) FullHelp() []string {
	r := []string{
		"$> ffsclient tabs (list)",
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

func (a *CLIArgumentsTabsBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient tabs must be called with a subcommand (eg `ffsclient tabs list`)")
}

func (a *CLIArgumentsTabsBase) Execute(ctx *cli.FFSContext) error {
	return fferr.NewDirectOutput(consts.ExitcodeError, "Cannot call `tabs` command without an subcommand")
}

type CLIArgumentsTabsUtil struct {
	CLIArgumentsBaseUtil
}

func (a *CLIArgumentsTabsUtil) filterDeletedSingle(ctx *cli.FFSContext, records []models.TabRecord, includeDeleted bool, onlyDeleted bool, client *[]string) []models.TabRecord {
	result := make([]models.TabRecord, 0, len(records))

	for _, v := range records {
		if v.ClientDeleted && !includeDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v[%d] (is deleted and include-deleted == false)", v.ClientID, v.Index))
			continue
		}

		if !v.ClientDeleted && onlyDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v[%d] (is not deleted and only-deleted == true)", v.ClientID, v.Index))
			continue
		}

		if client != nil && !langext.InArray(v.ClientID, *client) {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v[%d] (not in client-filter)", v.ClientID, v.Index))
			continue
		}

		result = append(result, v)
	}

	return result
}

func (a *CLIArgumentsTabsUtil) filterDeletedMulti(ctx *cli.FFSContext, records []models.TabClientRecord, includeDeleted bool, onlyDeleted bool, client *[]string) []models.TabClientRecord {
	result := make([]models.TabClientRecord, 0, len(records))

	for _, v := range records {
		if v.Deleted && !includeDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is deleted and include-deleted == false)", v.ID))
			continue
		}

		if !v.Deleted && onlyDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is not deleted and only-deleted == true)", v.ID))
			continue
		}

		if client != nil && !langext.InArray(v.ID, *client) {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (not in client-filter)", v.ID))
			continue
		}

		result = append(result, v)
	}

	return result
}

func (a *CLIArgumentsTabsUtil) LastHistory(v models.TabRecord, def string) string {
	if len(v.UrlHistory) == 0 {
		return def
	}

	return v.UrlHistory[len(v.UrlHistory)-1]
}
