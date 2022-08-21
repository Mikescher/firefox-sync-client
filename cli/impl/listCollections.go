package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsListCollections struct {
}

func NewCLIArgumentsListCollections() *CLIArgumentsListCollections {
	return &CLIArgumentsListCollections{}
}

func (a *CLIArgumentsListCollections) Mode() cli.Mode {
	return cli.ModeListCollections
}

func (a *CLIArgumentsListCollections) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + positionalArgs[0])
	}

	if len(optionArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + optionArgs[0].Key)
	}

	return nil
}

func (a *CLIArgumentsListCollections) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[List collections]")
	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)

	cfp, err := ctx.AbsConfigFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !langext.FileExists(cfp) {
		ctx.PrintFatalMessage("Configfile does not exist.")
		ctx.PrintFatalMessage("Use `ffsclient login <email> <password>` first")
		return consts.ExitcodeNoLogin
	}

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	collections, err := client.ListCollections(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	panic(0) //TODO
}
