package impl

import (
	"ffsyncclient/cli"
	"regexp"
	"strings"
)

var urlSchemaRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]+://`)

func ParseSubcommand(v []string) (cli.Verb, string, int, bool) {

	idx := 0

	if len(v) == 0 || strings.HasPrefix(v[0], "-") {
		return nil, "", 0, false
	}

	for idx < len(v) && !strings.HasPrefix(v[idx], "-") {
		idx++
	}

	for i := idx; i > 0; i-- {

		verb := strings.Join(v[:i], " ")

		if sc, ok := GetSubcommand(verb); ok {
			return sc, verb, i, true
		}

	}

	return nil, v[0], 1, false

}

func GetSubcommand(v string) (cli.Verb, bool) {
	for _, verb := range cli.Modes {
		if strings.ToLower(v) == strings.ToLower(string(verb)) {
			return GetModeImpl(verb), true
		}
	}

	return nil, false
}

func GetModeImpl(m cli.Mode) cli.Verb {
	switch m {

	case cli.ModeHelp:
		return NewCLIArgumentsHelp()
	case cli.ModeVersion:
		return NewCLIArgumentsVersion()
	case cli.ModeLogin:
		return NewCLIArgumentsLogin()
	case cli.ModeTokenRefresh:
		return NewCLIArgumentsTokenRefresh()
	case cli.ModeCheckSession:
		return NewCLIArgumentsCheckSession()
	case cli.ModeQuotaGet:
		return NewCLIArgumentsQuotaGet()
	case cli.ModeCollectionsList:
		return NewCLIArgumentsCollectionsList()
	case cli.ModeRecordsList:
		return NewCLIArgumentsRecordsList()
	case cli.ModeRecordsGet:
		return NewCLIArgumentsRecordsGet()
	case cli.ModeRecordsDelete:
		return NewCLIArgumentsRecordsDelete()
	case cli.ModeCollectionsDelete:
		return NewCLIArgumentsCollectionsDelete()
	case cli.ModeDeleteAll:
		return NewCLIArgumentsDeleteAll()
	case cli.ModeRecordsCreate:
		return NewCLIArgumentsRecordsCreate()
	case cli.ModeRecordsUpdate:
		return NewCLIArgumentsRecordsUpdate()
	case cli.ModeMetaGet:
		return NewCLIArgumentsMetaGet()
	case cli.ModeBookmarksBase:
		return NewCLIArgumentsBookmarksBase()
	case cli.ModeBookmarksList:
		return NewCLIArgumentsBookmarksList()
	case cli.ModeBookmarksDelete:
		return NewCLIArgumentsBookmarksDelete()
	case cli.ModeBookmarksCreateBase:
		return NewCLIArgumentsBookmarksCreateBase()
	case cli.ModeBookmarksCreateBookmark:
		return NewCLIArgumentsBookmarksCreateBookmark()
	case cli.ModeBookmarksCreateFolder:
		return NewCLIArgumentsBookmarksCreateFolder()
	case cli.ModeBookmarksCreateSeparator:
		return NewCLIArgumentsBookmarksCreateSeparator()
	case cli.ModeBookmarksUpdate:
		return NewCLIArgumentsBookmarksUpdate()
	case cli.ModePasswordsBase:
		return NewCLIArgumentsPasswordsBase()
	case cli.ModePasswordsList:
		return NewCLIArgumentsPasswordsList()
	case cli.ModePasswordsDelete:
		return NewCLIArgumentsPasswordsDelete()
	case cli.ModePasswordsCreate:
		return NewCLIArgumentsPasswordsCreate()
	case cli.ModePasswordsUpdate:
		return NewCLIArgumentsPasswordsUpdate()
	case cli.ModePasswordsGet:
		return NewCLIArgumentsPasswordsGet()
	case cli.ModeFormsBase:
		return NewCLIArgumentsFormsBase()
	case cli.ModeFormsList:
		return NewCLIArgumentsFormsList()
	case cli.ModeFormsGet:
		return NewCLIArgumentsFormsGet()
	case cli.ModeFormsCreate:
		return NewCLIArgumentsFormsCreate()
	case cli.ModeFormsUpdate:
		return NewCLIArgumentsFormsUpdate()
	case cli.ModeFormsDelete:
		return NewCLIArgumentsFormsDelete()
	case cli.ModeHistoryBase:
		return NewCLIArgumentsHistoryBase()
	case cli.ModeHistoryList:
		return NewCLIArgumentsHistoryList()
	case cli.ModeHistoryCreate:
		return NewCLIArgumentsHistoryCreate()
	case cli.ModeHistoryUpdate:
		return NewCLIArgumentsHistoryUpdate()
	case cli.ModeHistoryDelete:
		return NewCLIArgumentsHistoryDelete()

	default:
		panic("Unknown Mode: " + m)
	}
}

func ListSubcommands(m cli.Mode, skipBase bool) []cli.Mode {
	r := make([]cli.Mode, 0)

	for _, v := range cli.Modes {
		if strings.HasPrefix(string(v), string(m)+" ") {
			if skipBase && GetModeImpl(v).ShortHelp() == nil {
				continue
			}
			r = append(r, v)
		}
	}

	return r
}
