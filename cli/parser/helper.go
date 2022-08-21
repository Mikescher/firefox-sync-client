package parser

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/impl"
	"strings"
)

func getVerb(v string) (cli.Verb, bool) {
	switch strings.ToLower(v) {

	case strings.ToLower(string(cli.ModeHelp)):
		return impl.NewCLIArgumentsHelp(), true

	case strings.ToLower(string(cli.ModeVersion)):
		return impl.NewCLIArgumentsVersion(), true

	case strings.ToLower(string(cli.ModeLogin)):
		return impl.NewCLIArgumentsLogin(), true

	case strings.ToLower(string(cli.ModeDeleteAll)):
		return impl.NewCLIArgumentsDeleteAll(), true

	case strings.ToLower(string(cli.ModeDeleteSingle)):
		return impl.NewCLIArgumentsDeleteSingle(), true

	case strings.ToLower(string(cli.ModeListCollections)):
		return impl.NewCLIArgumentsListCollections(), true

	case strings.ToLower(string(cli.ModeGetRawRecord)):
		return impl.NewCLIArgumentsGetRawRecord(), true

	case strings.ToLower(string(cli.ModeGetDecodedRecord)):
		return impl.NewCLIArgumentsGetDecodedRecord(), true

	case strings.ToLower(string(cli.ModeCreateRecord)):
		return impl.NewCLIArgumentsCreateRecord(), true

	case strings.ToLower(string(cli.ModeUpdateRecord)):
		return impl.NewCLIArgumentsUpdateRecord(), true

	case strings.ToLower(string(cli.ModeGetQuota)):
		return impl.NewCLIArgumentsGetQuota(), true

	}

	return nil, false
}
