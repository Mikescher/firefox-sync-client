package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsRecordsUtil struct {
	CLIArgumentsBaseUtil
}

func (a *CLIArgumentsRecordsUtil) prettyPrint(ctx *cli.FFSContext, pretty bool, v string, xmlident bool) string {
	if pretty {
		pp, ok := langext.PrettyPrintJson(v)
		if !ok {
			return pp
		} else {
			if xmlident {
				return "\n" + langext.Indent(pp, "    ") + "  "
			} else {
				return pp
			}
		}
	} else {
		return v
	}
}

func (a *CLIArgumentsRecordsUtil) tryParseJson(ctx *cli.FFSContext, v []byte) any {
	var err error

	var j1 langext.H
	err = json.Unmarshal(v, &j1)
	if err == nil {
		return j1
	}

	var j2 langext.A
	err = json.Unmarshal(v, &j2)
	if err == nil {
		return j2
	}

	return v
}
