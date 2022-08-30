package cli

type Mode string

const (
	ModeHelp              Mode = "help"
	ModeVersion           Mode = "version"
	ModeLogin             Mode = "login"
	ModeTokenRefresh      Mode = "refresh"
	ModeCheckSession      Mode = "check-session"
	ModeQuotaGet          Mode = "quota"
	ModeCollectionsList   Mode = "collections"
	ModeRecordsList       Mode = "list"
	ModeRecordsGet        Mode = "get"
	ModeRecordsDelete     Mode = "delete"
	ModeCollectionsDelete Mode = "delete-collection"
	ModeDeleteAll         Mode = "delete-all"
	ModeRecordsCreate     Mode = "create"
	ModeRecordsUpdate     Mode = "update"
	ModeMetaGet           Mode = "meta"
	ModeBookmarksBase     Mode = "bookmarks"        // TODO
	ModeBookmarksList     Mode = "bookmarks list"   // TODO
	ModeBookmarksDelete   Mode = "bookmarks delete" // TODO
	ModeBookmarksCreate   Mode = "bookmarks create" // TODO
	ModeBookmarksUpdate   Mode = "bookmarks update" // TODO
	ModePasswordsBase     Mode = "passwords"
	ModePasswordsList     Mode = "passwords list"
	ModePasswordsGet      Mode = "passwords get"
	ModePasswordsCreate   Mode = "passwords create"
	ModePasswordsUpdate   Mode = "passwords update" // TODO
	ModePasswordsDelete   Mode = "passwords delete"
	ModeFormsBase         Mode = "forms"          // TODO
	ModeFormsList         Mode = "forms list"     // TODO
	ModeFormsGet          Mode = "forms get"      // TODO
	ModeFormsCreate       Mode = "forms create"   // TODO
	ModeFormsUpdate       Mode = "forms update"   // TODO
	ModeFormsDelete       Mode = "forms delete"   // TODO
	ModeHistoryBase       Mode = "history"        // TODO
	ModeHistoryList       Mode = "history list"   // TODO
	ModeHistoryCreate     Mode = "history create" // TODO
	ModeHistoryUpdate     Mode = "history update" // TODO
	ModeHistoryDelete     Mode = "history delete" // TODO
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
	PositionArgCount() (*int, *int)
}
