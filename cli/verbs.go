package cli

type Mode string

const (
	ModeHelp            Mode = "help"
	ModeVersion         Mode = "version"
	ModeLogin           Mode = "login"
	ModeTokenRefresh    Mode = "refresh"
	ModeCheckSession    Mode = "check-session"
	ModeGetQuota        Mode = "quota"
	ModeListCollections Mode = "collections"
	ModeListRecords     Mode = "list"
	ModeGetRecord       Mode = "get"
	ModeDeleteAll       Mode = "delete-all" // TODO
	ModeDeleteSingle    Mode = "delete"     // TODO
	ModeCreateRecord    Mode = "create"     // TODO
	ModeUpdateRecord    Mode = "update"     // TODO
	//ModeBookmarks       Mode = "bookmarks"  // TODO
	//ModePasswords       Mode = "passwords"  // TODO
	//ModeForms           Mode = "forms"      // TODO
	//ModeHistory         Mode = "history"    // TODO
)

var Modes = []Mode{
	ModeLogin,
	ModeTokenRefresh,
	ModeCheckSession,

	ModeListCollections,
	ModeGetQuota,
	ModeListRecords,
	ModeGetRecord,
	//ModeDeleteSingle,
	//ModeDeleteAll,
	//ModeCreateRecord,
	//ModeUpdateRecord,

	//ModeBookmarks,
	//ModePasswords,
	//ModeForms,
	//ModeHistory,

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
