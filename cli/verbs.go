package cli

type Mode string

const (
	ModeHelp            Mode = "help"
	ModeVersion         Mode = "version"
	ModeLogin           Mode = "login"
	ModeTokenRefresh    Mode = "refresh"
	ModeGetQuota        Mode = "quota"
	ModeListCollections Mode = "collections"
	ModeDeleteAll       Mode = "delete-all" // TODO
	ModeDeleteSingle    Mode = "delete"     // TODO
	ModeGetRecord       Mode = "get"        // TODO
	ModeCreateRecord    Mode = "create"     // TODO
	ModeUpdateRecord    Mode = "update"     // TODO
	ModeListRecords     Mode = "list"       // TODO
	ModeBookmarks       Mode = "bookmarks"  // TODO
	ModePasswords       Mode = "passwords"  // TODO
	ModeForms           Mode = "forms"      // TODO
	ModeHistory         Mode = "history"    // TODO
)

func (m Mode) String() string {
	return string(m)
}

type Verb interface {
	Mode() Mode
	Init(positionalArgs []string, optionArgs []ArgumentTuple) error
	Execute(ctx *FFSContext) int
}
