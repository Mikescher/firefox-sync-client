package cli

type Mode string

const (
	ModeHelp                     Mode = "help"
	ModeVersion                  Mode = "version"
	ModeLogin                    Mode = "login"
	ModeTokenRefresh             Mode = "refresh"
	ModeCheckSession             Mode = "check-session"
	ModeQuotaGet                 Mode = "quota"
	ModeCollectionsList          Mode = "collections"
	ModeRecordsList              Mode = "list"
	ModeRecordsGet               Mode = "get"
	ModeRecordsDelete            Mode = "delete"
	ModeCollectionsDelete        Mode = "delete-collection"
	ModeDeleteAll                Mode = "delete-all"
	ModeRecordsCreate            Mode = "create"
	ModeRecordsUpdate            Mode = "update"
	ModeMetaGet                  Mode = "meta"
	ModeBookmarksBase            Mode = "bookmarks"
	ModeBookmarksList            Mode = "bookmarks list"
	ModeBookmarksDelete          Mode = "bookmarks delete"
	ModeBookmarksCreateBase      Mode = "bookmarks create"
	ModeBookmarksCreateBookmark  Mode = "bookmarks create bookmark"
	ModeBookmarksCreateFolder    Mode = "bookmarks create folder"
	ModeBookmarksCreateSeparator Mode = "bookmarks create separator"
	ModeBookmarksUpdate          Mode = "bookmarks update"
	ModePasswordsBase            Mode = "passwords"
	ModePasswordsList            Mode = "passwords list"
	ModePasswordsGet             Mode = "passwords get"
	ModePasswordsCreate          Mode = "passwords create"
	ModePasswordsUpdate          Mode = "passwords update"
	ModePasswordsDelete          Mode = "passwords delete"
	ModeFormsBase                Mode = "forms"
	ModeFormsList                Mode = "forms list"
	ModeFormsGet                 Mode = "forms get"    // TODO
	ModeFormsCreate              Mode = "forms create" // TODO
	ModeFormsDelete              Mode = "forms delete" // TODO
	ModeHistoryBase              Mode = "history"
	ModeHistoryList              Mode = "history list"   // TODO
	ModeHistoryCreate            Mode = "history create" // TODO
	ModeHistoryDelete            Mode = "history delete" // TODO
)

var Modes = []Mode{
	ModeLogin,
	ModeTokenRefresh,
	ModeCheckSession,

	ModeCollectionsList,
	ModeQuotaGet,
	ModeRecordsList,
	ModeRecordsGet,
	ModeRecordsDelete,
	ModeCollectionsDelete,
	ModeDeleteAll,
	ModeRecordsCreate,
	ModeRecordsUpdate,
	ModeMetaGet,

	ModeBookmarksBase,
	ModeBookmarksList,
	ModeBookmarksDelete,
	ModeBookmarksCreateBase,
	ModeBookmarksCreateBookmark,
	ModeBookmarksCreateFolder,
	ModeBookmarksCreateSeparator,
	ModeBookmarksUpdate,

	ModePasswordsBase,
	ModePasswordsList,
	ModePasswordsDelete,
	ModePasswordsCreate,
	ModePasswordsUpdate,
	ModePasswordsGet,

	ModeFormsBase,
	ModeFormsList,
	ModeFormsGet,
	ModeFormsCreate,
	ModeFormsDelete,

	ModeHistoryBase,
	ModeHistoryList,
	ModeHistoryCreate,
	ModeHistoryDelete,

	ModeVersion,
	ModeHelp,
}

func (m Mode) String() string {
	return string(m)
}

type Verb interface {
	Mode() Mode
	Init(positionalArgs []string, optionArgs []ArgumentTuple) error
	Execute(ctx *FFSContext) int
	ShortHelp() [][]string
	FullHelp() []string
	PositionArgCount() (*int, *int)
}
