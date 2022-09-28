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

func (a *CLIArgumentsFormsGet) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Get Forms]")
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

	records, err := client.ListRecords(ctx, session, consts.CollectionForms, nil, nil, false, true, nil, nil)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	forms, err := models.UnmarshalForms(ctx, records, true)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
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

func (a *CLIArgumentsFormsGet) printOutput(ctx *cli.FFSContext, forms []models.FormRecord) int {

	switch langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) {

	case cli.OutputFormatTable:
		for _, v := range forms {
			ctx.PrintPrimaryOutput(v.Value)
		}
		return 0

	case cli.OutputFormatText:
		for _, v := range forms {
			ctx.PrintPrimaryOutput(v.Value)
		}
		return 0

	case cli.OutputFormatJson:
		json := langext.A{}
		for _, v := range forms {
			json = append(json, v.Value)
		}
		ctx.PrintPrimaryOutputJSON(json)
		return 0

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
		return 0

	default:
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}
}
