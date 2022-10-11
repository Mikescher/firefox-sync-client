package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"fmt"
)

type CLIArgumentsQuotaGet struct {
}

func NewCLIArgumentsQuotaGet() *CLIArgumentsQuotaGet {
	return &CLIArgumentsQuotaGet{}
}

func (a *CLIArgumentsQuotaGet) Mode() cli.Mode {
	return cli.ModeQuotaGet
}

func (a *CLIArgumentsQuotaGet) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsQuotaGet) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
}

func (a *CLIArgumentsQuotaGet) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient quota", "Query the storage quota of the current user"},
	}
}

func (a *CLIArgumentsQuotaGet) FullHelp() []string {
	return []string{
		"$> ffsclient quota",
		"",
		"Get the storage quota of the current user (used / max)",
	}
}

func (a *CLIArgumentsQuotaGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsQuotaGet) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Get Quota]")
	ctx.PrintVerbose("")

	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return err
	}

	if !langext.FileExists(cfp) {
		return fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return err
	}

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		return err
	}

	// ========================================================================

	used, total, err := client.GetQuota(ctx, session)
	if err != nil {
		return err
	}

	// ========================================================================

	return a.printOutput(ctx, total, used)
}

func (a *CLIArgumentsQuotaGet) printOutput(ctx *cli.FFSContext, total *int64, used int64) error {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatTable:
		if total == nil {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v    %v", langext.FormatBytes(used), "INF"))
		} else {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v    %v", langext.FormatBytes(used), langext.FormatBytes(*total)))
		}
		return nil

	case cli.OutputFormatText:
		if total == nil {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v / %v", langext.FormatBytes(used), "INF"))
		} else {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v / %v", langext.FormatBytes(used), langext.FormatBytes(*total)))
		}
		return nil

	case cli.OutputFormatJson:
		if total == nil {
			ctx.PrintPrimaryOutputJSON(langext.H{
				"used":        langext.FormatBytes(used),
				"used_bytes":  used,
				"total":       nil,
				"total_bytes": nil,
			})
		} else {
			ctx.PrintPrimaryOutputJSON(langext.H{
				"used":        langext.FormatBytes(used),
				"used_bytes":  used,
				"total":       langext.FormatBytes(*total),
				"total_bytes": *total,
			})
		}
		return nil

	case cli.OutputFormatXML:
		if total == nil {
			type xml struct {
				Used      string   `xml:"Used,omitempty,attr"`
				UsedBytes int64    `xml:"UsedBytes,omitempty,attr"`
				XMLName   struct{} `xml:"Quota"`
			}
			ctx.PrintPrimaryOutputXML(xml{
				Used:      langext.FormatBytes(used),
				UsedBytes: used,
			})
		} else {
			type xml struct {
				Used       string   `xml:"Used,omitempty,attr"`
				UsedBytes  int64    `xml:"UsedBytes,omitempty,attr"`
				Total      string   `xml:"Total,omitempty,attr"`
				TotalBytes int64    `xml:"TotalBytes,omitempty,attr"`
				XMLName    struct{} `xml:"Quota"`
			}
			ctx.PrintPrimaryOutputXML(xml{
				Used:       langext.FormatBytes(used),
				UsedBytes:  used,
				Total:      langext.FormatBytes(*total),
				TotalBytes: *total,
			})
		}
		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
