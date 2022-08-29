package cli

type Mode string

const (
	ModeHelp             Mode = "help"
	ModeVersion          Mode = "version"
	ModeLogin            Mode = "login"
	ModeTokenRefresh     Mode = "refresh"
	ModeCheckSession     Mode = "check-session"
	ModeGetQuota         Mode = "quota"
	ModeListCollections  Mode = "collections"
	ModeListRecords      Mode = "list"
	ModeGetRecord        Mode = "get"
	ModeDeleteRecord     Mode = "delete"
	ModeDeleteCollection Mode = "delete-collection"
	ModeDeleteAll        Mode = "delete-all"
	ModeCreateRecord     Mode = "create"           // TODO
	ModeUpdateRecord     Mode = "update"           // TODO
	ModeMeta             Mode = "meta"             // TODO
	ModeBookmarksBase    Mode = "bookmarks"        // TODO
	ModeBookmarksList    Mode = "bookmarks list"   // TODO
	ModeBookmarksDelete  Mode = "bookmarks delete" // TODO
	ModeBookmarksCreate  Mode = "bookmarks create" // TODO
	ModeBookmarksUpdate  Mode = "bookmarks update" // TODO
	ModePasswordsBase    Mode = "passwords"        // TODO
	ModePasswordsList    Mode = "passwords list"   // TODO
	ModePasswordsDelete  Mode = "passwords delete" // TODO
	ModePasswordsCreate  Mode = "passwords create" // TODO
	ModePasswordsUpdate  Mode = "passwords update" // TODO
	ModePasswordsGet     Mode = "passwords get"    // TODO
	ModeFormsBase        Mode = "forms"            // TODO
	ModeFormsList        Mode = "forms list"       // TODO
	ModeFormsGet         Mode = "forms get"        // TODO
	ModeFormsCreate      Mode = "forms create"     // TODO
	ModeFormsUpdate      Mode = "forms update"     // TODO
	ModeFormsDelete      Mode = "forms delete"     // TODO
	ModeHistoryBase      Mode = "history"          // TODO
	ModeHistoryList      Mode = "history list"     // TODO
	ModeHistoryCreate    Mode = "history create"   // TODO
	ModeHistoryUpdate    Mode = "history update"   // TODO
	ModeHistoryDelete    Mode = "history delete"   // TODO
)

var Modes = []Mode{
	ModeLogin,
	ModeTokenRefresh,
	ModeCheckSession,

	ModeListCollections,
	ModeGetQuota,
	ModeListRecords,
	ModeGetRecord,
	ModeDeleteRecord,
	ModeDeleteCollection,
	ModeDeleteAll,
	ModeCreateRecord,
	ModeUpdateRecord,
	ModeMeta,

	ModeBookmarksBase,
	ModeBookmarksList,
	ModeBookmarksDelete,
	ModeBookmarksCreate,
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
	ModeFormsUpdate,
	ModeFormsDelete,

	ModeHistoryBase,
	ModeHistoryList,
	ModeHistoryCreate,
	ModeHistoryUpdate,
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
}
