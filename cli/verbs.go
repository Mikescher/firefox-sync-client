package cli

type Mode string

const (
	ModeHelp            Mode = "help"
	ModeVersion         Mode = "version"
	ModeLogin           Mode = "login"
	ModeDeleteAll       Mode = "delete-all"
	ModeDeleteSingle    Mode = "delete"
	ModeListCollections Mode = "collections"
	ModeGetRecord       Mode = "get"
	ModeCreateRecord    Mode = "create"
	ModeUpdateRecord    Mode = "update"
)

type Verb interface {
	Mode() Mode
	Init(positionalArgs []string, optionArgs []ArgumentTuple) error
	Execute(ctx *FFSContext) int
}
