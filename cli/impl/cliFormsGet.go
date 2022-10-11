package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
	"strings"
)

type CLIArgumentsFormsGet struct {
	Name       string
	IgnoreCase bool

	CLIArgumentsFormsUtil
}

func NewCLIArgumentsFormsGet() *CLIArgumentsFormsGet {
	return &CLIArgumentsFormsGet{}
}

func (a *CLIArgumentsFormsGet) Mode() cli.Mode {
	return cli.ModeFormsGet
}

func (a *CLIArgumentsFormsGet) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsFormsGet) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatTable, cli.OutputFormatText, cli.OutputFormatJson, cli.OutputFormatXML}
}

func (a *CLIArgumentsFormsGet) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient forms get <name> [--ignore-case]", "Get all HTML-Form autocomplete suggestions for this name"},
	}
}

func (a *CLIArgumentsFormsGet) FullHelp() []string {
	return []string{
		"$> ffsclient forms get <name>",
		"",
		"Get all HTML-Form autocomplete suggestions for this name.",
		"The name is matched (by default) case-sensitive. Use case-insensitive comparison with the --ignore-case flag",
	}
}

func (a *CLIArgumentsFormsGet) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Name = positionalArgs[0]

	for _, arg := range optionArgs {
		if arg.Key == "ignore-case" && arg.Value == nil {
			a.IgnoreCase = true
			continue
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsFormsGet) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Get Forms]")
	ctx.PrintVerbose("")

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

	records, err := client.ListRecords(ctx, session, consts.CollectionForms, nil, nil, false, true, nil, nil)
	if err != nil {
		return err
	}

	forms, err := models.UnmarshalForms(ctx, records, true)
	if err != nil {
		return err
	}

	filteredForms := make([]models.FormRecord, 0, len(forms))
	for _, v := range forms {
		if v.Deleted {
			continue
		}
		if v.Name == a.Name {
			filteredForms = append(filteredForms, v)
			ctx.PrintVerbose(fmt.Sprintf("Entry %s does match case-sensitive name-filter ('%s' === '%s')", v.ID, a.Name, v.Name))
			continue
		}
		if a.IgnoreCase && strings.EqualFold(v.Name, a.Name) {
			filteredForms = append(filteredForms, v)
			ctx.PrintVerbose(fmt.Sprintf("Entry %s does match case-insensitive name-filter ('%s' =~= '%s')", v.ID, a.Name, v.Name))
			continue
		}

		ctx.PrintVerbose(fmt.Sprintf("Entry %s does not match name-filter ('%s' <> '%s')", v.ID, a.Name, v.Name))
	}

	// ========================================================================

	return a.printOutput(ctx, filteredForms)
}

func (a *CLIArgumentsFormsGet) printOutput(ctx *cli.FFSContext, forms []models.FormRecord) error {

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatTable:
		for _, v := range forms {
			ctx.PrintPrimaryOutput(v.Value)
		}
		return nil

	case cli.OutputFormatText:
		for _, v := range forms {
			ctx.PrintPrimaryOutput(v.Value)
		}
		return nil

	case cli.OutputFormatJson:
		json := langext.A{}
		for _, v := range forms {
			json = append(json, v.Value)
		}
		ctx.PrintPrimaryOutputJSON(json)
		return nil

	case cli.OutputFormatXML:
		type xmlroot struct {
			Entries []string `xml:"form"`
			XMLName struct{} `xml:"values"`
		}
		node := xmlroot{Entries: make([]string, 0, len(forms))}
		for _, v := range forms {
			node.Entries = append(node.Entries, v.Value)
		}
		ctx.PrintPrimaryOutputXML(node)
		return nil

	default:
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}
}
