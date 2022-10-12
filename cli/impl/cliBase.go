package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/syncclient"
)

type CLIArgumentsBaseUtil struct{}

func (a *CLIArgumentsBaseUtil) InitClient(ctx *cli.FFSContext) (*syncclient.FxAClient, syncclient.FFSyncSession, error) {
	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return nil, syncclient.FFSyncSession{}, err
	}

	if !langext.FileExists(cfp) {
		return nil, syncclient.FFSyncSession{}, fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return nil, syncclient.FFSyncSession{}, err
	}

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		return nil, syncclient.FFSyncSession{}, err
	}

	return client, session, nil
}
