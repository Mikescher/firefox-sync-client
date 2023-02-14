package cli

import (
	"strings"
)

type OutputFormat string

const (
	OutputFormatText     OutputFormat = "text"
	OutputFormatJson     OutputFormat = "json"
	OutputFormatNetscape OutputFormat = "netscape"
	OutputFormatXML      OutputFormat = "xml"
	OutputFormatTable    OutputFormat = "table"
	OutputFormatTSV      OutputFormat = "tsv"
)

func (f OutputFormat) String() string {
	return string(f)
}

func GetOutputFormat(v string) (OutputFormat, bool) {
	switch strings.ToLower(v) {

	case strings.ToLower(string(OutputFormatText)):
		return OutputFormatText, true

	case strings.ToLower(string(OutputFormatJson)):
		return OutputFormatJson, true

	case strings.ToLower(string(OutputFormatNetscape)):
		return OutputFormatNetscape, true

	case strings.ToLower(string(OutputFormatXML)):
		return OutputFormatXML, true

	case strings.ToLower(string(OutputFormatTable)):
		return OutputFormatTable, true

	case strings.ToLower(string(OutputFormatTSV)):
		return OutputFormatTSV, true

	default:
		return "", false
	}

}
