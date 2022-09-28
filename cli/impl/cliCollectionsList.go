package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
	"strconv"
)

type CLIArgumentsCollectionsList struct {
	ShowUsage bool
}

func NewCLIArgumentsCollectionsList() *CLIArgumentsCollectionsList {
	return &CLIArgumentsCollectionsList{
		ShowUsage: false,
	}
}

func (a *CLIArgumentsCollectionsList) Mode() cli.Mode {
	return cli.ModeCollectionsList
}

func (a *CLIArgumentsCollectionsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsCollectionsList) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
}

func (a *CLIArgumentsCollectionsList) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient collections", "List all available collections"},
		{"          [--usage]", "Include usage (storage space)"},
	}
}

func (a *CLIArgumentsCollectionsList) FullHelp() []string {
	return []string{
		"$> ffsclient collections [--usage]",
		"",
		"List all available collections together with their last-modified-time and entry-count",
		"",
		"Optionally includes the storage-space usage (Note: This request may be very expensive)",
	}
}

func (a *CLIArgumentsCollectionsList) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	for _, arg := range optionArgs {
		if arg.Key == "usage" && arg.Value == nil {
			a.ShowUsage = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsCollectionsList) Execute(ctx *cli.FFSContext) int {
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
		return -1, fferr.DirectOutput.New("Collection '" + needle + "' not found")
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

func (a *CLIArgumentsCollectionsList) printOutput(ctx *cli.FFSContext, collections []models.Collection) int {
	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable) {
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

		type xmlentry struct {
			Name       string `xml:",chardata"`
			Time       string `xml:"LastModified,attr"`
			TimeUnix   string `xml:"LastModifiedUnix,attr"`
			Count      string `xml:"Count,attr"`
			Usage      string `xml:"Usage,omitempty,attr"`
			UsageBytes string `xml:"UsageBytes,omitempty,attr"`
		}
		type xml struct {
			Collections []xmlentry `xml:"Collection"`
			XMLName     struct{}   `xml:"Collections"`
		}

		node := xml{Collections: make([]xmlentry, 0, len(collections))}
		for _, v := range collections {
			obj := xmlentry{
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

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}
