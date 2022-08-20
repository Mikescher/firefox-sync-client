package cli

type Mode string

const (
	ModeHelp             Mode = "help"
	ModeVersion          Mode = "version"
	ModeLogin            Mode = "login"
	ModeDeleteAll        Mode = "delete-all"
	ModeDeleteSingle     Mode = "delete"
	ModeListCollections  Mode = "collections"
	ModeGetRawRecord     Mode = "raw"
	ModeGetDecodedRecord Mode = "get"
	ModeCreateRecord     Mode = "create"
	ModeUpdateRecord     Mode = "update"
)

func (m Mode) String() string {
	return string(m)
}

type Verb interface {
	Mode() Mode
	Init(positionalArgs []string, optionArgs []ArgumentTuple) error
	Execute(ctx *FFSContext) int
}
