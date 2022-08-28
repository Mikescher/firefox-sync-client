package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"fmt"
	"github.com/joomcode/errorx"
)

type CLIArgumentsGetQuota struct {
}

func NewCLIArgumentsGetQuota() *CLIArgumentsGetQuota {
	return &CLIArgumentsGetQuota{}
}

func (a *CLIArgumentsGetQuota) Mode() cli.Mode {
	return cli.ModeGetQuota
}

func (a *CLIArgumentsGetQuota) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsGetQuota) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Get Quota]")
	ctx.PrintVerbose("")

	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !langext.FileExists(cfp) {
		ctx.PrintFatalMessage("Sessionfile does not exist.")
		ctx.PrintFatalMessage("Use `ffsclient login <email> <password>` first")
		return consts.ExitcodeNoLogin
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	used, total, err := client.GetQuota(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	return a.printOutput(ctx, total, used)
}

func (a *CLIArgumentsGetQuota) printOutput(ctx *cli.FFSContext, total *int64, used int64) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatText:
		if total == nil {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v / %v", langext.FormatBytes(used), "INF"))
		} else {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v / %v", langext.FormatBytes(used), langext.FormatBytes(*total)))
		}
		return 0

	case cli.OutputFormatTable:
		if total == nil {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v    %v", langext.FormatBytes(used), "INF"))
		} else {
			ctx.PrintPrimaryOutput(fmt.Sprintf("%v    %v", langext.FormatBytes(used), langext.FormatBytes(*total)))
		}
		return 0

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
		return 0

	case cli.OutputFormatXML:
		if total == nil {
			type xmlcoll struct {
				Used      string   `xml:"Used,omitempty,attr"`
				UsedBytes int64    `xml:"UsedBytes,omitempty,attr"`
				XMLName   struct{} `xml:"Quota"`
			}
			ctx.PrintPrimaryOutputXML(xmlcoll{
				Used:      langext.FormatBytes(used),
				UsedBytes: used,
			})
		} else {
			type xmlcoll struct {
				Used       string   `xml:"Used,omitempty,attr"`
				UsedBytes  int64    `xml:"UsedBytes,omitempty,attr"`
				Total      string   `xml:"Total,omitempty,attr"`
				TotalBytes int64    `xml:"TotalBytes,omitempty,attr"`
				XMLName    struct{} `xml:"Quota"`
			}
			ctx.PrintPrimaryOutputXML(xmlcoll{
				Used:       langext.FormatBytes(used),
				UsedBytes:  used,
				Total:      langext.FormatBytes(*total),
				TotalBytes: *total,
			})
		}
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return 0
	}
}
