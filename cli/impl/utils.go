package impl

import (
	"ffsyncclient/cli"
	"strings"
)

func ParseVerb(v string) (cli.Verb, bool) {
	for _, verb := range cli.Modes {
		if strings.ToLower(v) == strings.ToLower(string(verb)) {
			return GetModeImpl(verb), true
		}
	}

	return nil, false
}

func GetModeImpl(v cli.Mode) cli.Verb {
	switch v {

	case cli.ModeHelp:
		return NewCLIArgumentsHelp()

	case cli.ModeVersion:
		return NewCLIArgumentsVersion()

	case cli.ModeLogin:
		return NewCLIArgumentsLogin()

	case cli.ModeDeleteAll:
		return NewCLIArgumentsDeleteAll()

	case cli.ModeDeleteRecord:
		return NewCLIArgumentsDeleteSingle()

	case cli.ModeDeleteCollection:
		return NewCLIArgumentsDeleteCollection()

	case cli.ModeListCollections:
		return NewCLIArgumentsListCollections()

	case cli.ModeGetRecord:
		return NewCLIArgumentsGetRecords()

	case cli.ModeCreateRecord:
		return NewCLIArgumentsCreateRecord()

	case cli.ModeUpdateRecord:
		return NewCLIArgumentsUpdateRecord()

	case cli.ModeGetQuota:
		return NewCLIArgumentsGetQuota()

	case cli.ModeTokenRefresh:
		return NewCLIArgumentsTokenRefresh()

	case cli.ModeListRecords:
		return NewCLIArgumentsListRecords()

	case cli.ModeCheckSession:
		return NewCLIArgumentsCheckSession()

	default:
		panic("Unknown Mode: " + v)
	}
}
