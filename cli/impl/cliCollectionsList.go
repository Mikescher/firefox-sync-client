package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"sort"
	"strconv"
)

type CLIArgumentsCollectionsList struct {
	ShowUsage bool
	Sorted    bool
	CLIArgumentsBaseUtil
}

func NewCLIArgumentsCollectionsList() *CLIArgumentsCollectionsList {
	return &CLIArgumentsCollectionsList{
		ShowUsage: false,
		Sorted:    true,
	}
}

func (a *CLIArgumentsCollectionsList) Mode() cli.Mode {
	return cli.ModeCollectionsList
}

func (a *CLIArgumentsCollectionsList) PositionArgCount() (*int, *int) {
	return langext.Ptr(0), langext.Ptr(0)
}

func (a *CLIArgumentsCollectionsList) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML, cli.OutputFormatCSV, cli.OutputFormatTSV}
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
		if arg.Key == "unsorted" && arg.Value == nil {
			a.Sorted = false
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsCollectionsList) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[List collections]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
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
		return err
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
		return err
	}
	for _, v := range collectionCounts {
		idx, err := search(collections, v.Name)
		if err != nil {
			return err
		}
		collections[idx].Count = v.Count
	}

	if a.ShowUsage {
		collectionUsages, err := client.GetCollectionsUsage(ctx, session)
		if err != nil {
			return err
		}
		for _, v := range collectionUsages {
			idx, err := search(collections, v.Name)
			if err != nil {
				return err
			}
			collections[idx].Usage = v.Usage
		}
	}

	if a.Sorted {
		sort.Slice(collections, func(i1, i2 int) bool {
			return collections[i1].Name < collections[i2].Name
		})
	}

	// ========================================================================

	return a.printOutput(ctx, collections)
}

func (a *CLIArgumentsCollectionsList) printOutput(ctx *cli.FFSContext, collections []models.Collection) error {

	ofmt := langext.Coalesce(ctx.Opt.Format, cli.OutputFormatTable)
	switch ofmt {
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

		ctx.PrintPrimaryOutputTable(table)
		return nil

	case cli.OutputFormatText:
		for _, v := range collections {
			if a.ShowUsage {
				ctx.PrintPrimaryOutput(fmt.Sprintf("%v %v %v %v", v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), v.Count, langext.FormatBytes(v.Usage)))
			} else {
				ctx.PrintPrimaryOutput(fmt.Sprintf("%v %v %v", v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), v.Count))
			}
		}
		return nil

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
		return nil

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
		return nil

	case cli.OutputFormatTSV:
		fallthrough
	case cli.OutputFormatCSV:
		table := make([][]string, 0, len(collections))
		if a.ShowUsage {
			table = append(table, []string{"Name", "LastModified", "Count", "Usage"})
		} else {
			table = append(table, []string{"Name", "LastModified", "Count"})
		}
		for _, v := range collections {
			if a.ShowUsage {
				table = append(table, []string{v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), strconv.Itoa(v.Count), langext.FormatBytes(v.Usage)})
			} else {
				table = append(table, []string{v.Name, v.LastModified.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat), strconv.Itoa(v.Count)})
			}
		}

		ctx.PrintPrimaryOutputCSV(table, ofmt == cli.OutputFormatTSV)

		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
