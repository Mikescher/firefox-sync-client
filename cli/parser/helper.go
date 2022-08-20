package parser

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/impl"
	"strings"
)

func getVerb(v string) (cli.Verb, bool) {
	switch strings.ToLower(v) {

	case strings.ToLower(string(cli.ModeHelp)):
		return &impl.CLIArgumentsHelp{}, true

	case strings.ToLower(string(cli.ModeVersion)):
		return &impl.CLIArgumentsVersion{}, true

	case strings.ToLower(string(cli.ModeLogin)):
		return &impl.CLIArgumentsLogin{}, true

	case strings.ToLower(string(cli.ModeDeleteAll)):
		return &impl.CLIArgumentsDeleteAll{}, true

	case strings.ToLower(string(cli.ModeDeleteSingle)):
		return &impl.CLIArgumentsDeleteSingle{}, true

	case strings.ToLower(string(cli.ModeListCollections)):
		return &impl.CLIArgumentsListCollections{}, true

	case strings.ToLower(string(cli.ModeGetRawRecord)):
		return &impl.CLIArgumentsGetRawRecord{}, true

	case strings.ToLower(string(cli.ModeGetDecodedRecord)):
		return &impl.CLIArgumentsGetDecodedRecord{}, true

	case strings.ToLower(string(cli.ModeCreateRecord)):
		return &impl.CLIArgumentsCreateRecord{}, true

	case strings.ToLower(string(cli.ModeUpdateRecord)):
		return &impl.CLIArgumentsUpdateRecord{}, true

	}

	return nil, false
}
