package cli

type Mode string

const (
	ModeHelp             Mode = "help"
	ModeVersion          Mode = "version"
	ModeLogin            Mode = "login"
	ModeTokenRefresh     Mode = "refresh"
	ModeGetQuota         Mode = "quota"
	ModeListCollections  Mode = "collections"
	ModeDeleteAll        Mode = "delete-all" // TODO
	ModeDeleteSingle     Mode = "delete"     // TODO
	ModeGetRawRecord     Mode = "raw"        // TODO
	ModeGetDecodedRecord Mode = "get"        // TODO
	ModeCreateRecord     Mode = "create"     // TODO
	ModeUpdateRecord     Mode = "update"     // TODO
	ModeGetAllRaw        Mode = "list-raw"   // TODO
	ModeGetAllDecoded    Mode = "list"       // TODO
)

func (m Mode) String() string {
	return string(m)
}

type Verb interface {
	Mode() Mode
	Init(positionalArgs []string, optionArgs []ArgumentTuple) error
	Execute(ctx *FFSContext) int
}
