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
	ctx.PrintVerbose("Server          := " + ctx.Opt.ServerURL)

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

	client := syncclient.NewFxAClient(ctx.Opt.ServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	sessionext, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	//TODO check_session_status

	sessionHawk, err := client.HawkAuth(ctx, sessionext)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	_, err = client.ListCollections(ctx, sessionHawk)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	panic(0) //TODO
}
