package cli

import (
	"strings"
)

type OutputFormat string //@enum:type

const (
	OutputFormatText     OutputFormat = "text"
	OutputFormatJson     OutputFormat = "json"
	OutputFormatNetscape OutputFormat = "netscape"
	OutputFormatXML      OutputFormat = "xml"
	OutputFormatTable    OutputFormat = "table"
	OutputFormatTSV      OutputFormat = "tsv"
	OutputFormatCSV      OutputFormat = "csv"
)

func GetOutputFormat(v string) (OutputFormat, bool) {
	for _, fmt := range OutputFormatValues() {
		if strings.EqualFold(v, fmt.String()) {
			return fmt, true
		}
	}

	return "", false
}
