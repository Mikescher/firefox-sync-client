package impl

import (
	"encoding/hex"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
)

type CLIArgumentsLogin struct {
	Email    string
	Password string
}

func (a *CLIArgumentsLogin) Mode() cli.Mode {
	return cli.ModeLogin
}

func (a *CLIArgumentsLogin) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) < 2 {
		return errorx.InternalError.New("Not enough arguments for <login>")
	}
	if len(positionalArgs) > 2 {
		return errorx.InternalError.New("Too many arguments for <login>")
	}
	if len(optionArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + optionArgs[0].Key)
	}

	a.Email = positionalArgs[0]
	a.Password = positionalArgs[1]

	return nil
}

func (a *CLIArgumentsLogin) Execute(ctx *cli.FFSContext) int {

	ctx.PrintVerbose("Login against endpoint " + ctx.Opt.ServerURL)
	ctx.PrintVerbose("Email           := " + a.Email)
	ctx.PrintVerbose("Password        := " + a.Password)

	cfp, err := ctx.AbsConfigFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	client := syncclient.NewFxAClient(ctx.Opt.ServerURL)

	session, err := client.Login(ctx, a.Email, a.Password)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Fetch session keys")

	keyA, keyB, err := client.FetchKeys(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Key[a]          := " + hex.EncodeToString(keyA))
	ctx.PrintVerbose("Key[b]          := " + hex.EncodeToString(keyB))

	ctx.PrintVerbose("Save session-config to " + ctx.Opt.ConfigFilePath)

	extsession := session.Extend(keyA, keyB)

	err = extsession.Save(cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Session saved")

	return 0
}
