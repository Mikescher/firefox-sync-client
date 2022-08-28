package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
	"github.com/joomcode/errorx"
	"strconv"
)

type CLIArgumentsListCollections struct {
	ShowUsage bool
}

func NewCLIArgumentsListCollections() *CLIArgumentsListCollections {
	return &CLIArgumentsListCollections{
		ShowUsage: false,
	}
}

func (a *CLIArgumentsListCollections) Mode() cli.Mode {
	return cli.ModeListCollections
}

func (a *CLIArgumentsListCollections) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient collections", "List all available collections"},
		{"          [--usage]", "Include usage (storage space)"},
	}
}

func (a *CLIArgumentsListCollections) FullHelp() []string {
	return []string{
		"$> ffsclient collections [--usage]",
		"",
		"List all available collections together with their last-modified-time and entry-count",
		"",
		"Optionally includes the storage-space usage (Note: This request may be very expensive)",
	}
}

func (a *CLIArgumentsListCollections) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	for _, arg := range optionArgs {
		if arg.Key == "usage" && arg.Value == nil {
			a.ShowUsage = true
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsListCollections) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[List collections]")
	ctx.PrintVerbose("")

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

	search := func(arr []models.Collection, needle string) (int, error) {
		for i, v := range arr {
			if v.Name == needle {
				return i, nil
			}
		}
		return -1, errorx.InternalError.New("collection '" + needle + "' not found")
	}

	collectionInfos, err := client.GetCollectionsInfo(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	collections := make([]models.Collection, 0, len(collectionInfos))
	for _, v := range collectionInfos {
		collections = append(collections, models.Collection{
			Name:         v.Name,
			LastModified: v.LastModified,
		})
	}

	collectionCounts, err := client.GetCollectionsCounts(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}
	for _, v := range collectionCounts {
		idx, err := search(collections, v.Name)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}
		collections[idx].Count = v.Count
	}

	if a.ShowUsage {
		collectionUsages, err := client.GetCollectionsUsage(ctx, session)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}
		for _, v := range collectionUsages {
			idx, err := search(collections, v.Name)
			if err != nil {
				ctx.PrintFatalError(err)
				return consts.ExitcodeError
			}
			collections[idx].Usage = v.Usage
		}
	}

	// ========================================================================

	return a.printOutput(ctx, collections)
}

func (a *CLIArgumentsListCollections) printOutput(ctx *cli.FFSContext, collections []models.Collection) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable) {
	case cli.OutputFormatText:
		for _, v := range collections {
			if a.ShowUsage {
				ctx.PrintPrimaryOutput(fmt.Sprintf("%v %v %v %v", v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), v.Count, langext.FormatBytes(v.Usage)))
			} else {
				ctx.PrintPrimaryOutput(fmt.Sprintf("%v %v %v", v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), v.Count))
			}
		}
		return 0

	case cli.OutputFormatJson:
		json := langext.A{}
		for _, v := range collections {
			obj := langext.H{
				"name":              v.Name,
				"lastModified":      v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
				"lastModified_unix": v.LastModified.Unix(),
				"count":             v.Count,
			}
			if a.ShowUsage {
				obj["usage"] = langext.FormatBytes(v.Usage)
				obj["usage_bytes"] = v.Usage
			}
			json = append(json, obj)
		}
		ctx.PrintPrimaryOutputJSON(json)
		return 0
	case cli.OutputFormatXML:

		type xmlcoll struct {
			Name       string `xml:"Name,attr"`
			Time       string `xml:"LastModified,attr"`
			TimeUnix   string `xml:"LastModifiedUnix,attr"`
			Count      string `xml:"Count,attr"`
			Usage      string `xml:"Usage,omitempty,attr"`
			UsageBytes string `xml:"UsageBytes,omitempty,attr"`
		}
		type xml struct {
			Collections []xmlcoll `xml:"Collection"`
			XMLName     struct{}  `xml:"Collections"`
		}

		node := xml{}
		for _, v := range collections {
			obj := xmlcoll{
				Name:     v.Name,
				Time:     v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
				TimeUnix: fmt.Sprintf("%d", v.LastModified.Unix()),
				Count:    fmt.Sprintf("%d", v.Count),
			}
			if a.ShowUsage {
				obj.Usage = langext.FormatBytes(v.Usage)
				obj.UsageBytes = fmt.Sprintf("%d", v.Usage)
			}
			node.Collections = append(node.Collections, obj)
		}
		ctx.PrintPrimaryOutputXML(node)
		return 0

	case cli.OutputFormatTable:
		table := make([][]string, 0, len(collections))
		if a.ShowUsage {
			table = append(table, []string{"NAME", "LAST MODIFIED", "COUNT", "USAGE"})
		} else {
			table = append(table, []string{"NAME", "LAST MODIFIED", "COUNT"})
		}
		for _, v := range collections {
			if a.ShowUsage {
				table = append(table, []string{v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), strconv.Itoa(v.Count), langext.FormatBytes(v.Usage)})
			} else {
				table = append(table, []string{v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), strconv.Itoa(v.Count)})
			}
		}

		ctx.PrintPrimaryOutputTable(table, true)
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}
