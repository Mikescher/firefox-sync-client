package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/syncclient"
	"fmt"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"os"
)

type CLIArgumentsBaseUtil struct{}

func (a *CLIArgumentsBaseUtil) InitClient(ctx *cli.FFSContext) (*syncclient.FxAClient, syncclient.FFSyncSession, error) {

	if ctx.Opt.ManualAuthLoginEmail != nil || ctx.Opt.ManualAuthLoginPassword != nil {

		ctx.PrintVerboseHeader("Do manual (temporary) login without session-storage")

		email := langext.Coalesce(ctx.Opt.ManualAuthLoginEmail, "")
		passw := langext.Coalesce(ctx.Opt.ManualAuthLoginPassword, "")

		ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
		ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)
		ctx.PrintVerboseKV("Email", email)
		ctx.PrintVerboseKV("Password", passw)

		hostname, err := os.Hostname()
		deviceName := "Firefox-Sync-Client (temp)"
		if err == nil {
			deviceName = "Firefox-Sync-Client (temp) on " + hostname
		}
		deviceType := "cli"

		client := syncclient.NewFxAClient(ctx, ctx.Opt.AuthServerURL)

		sessionCrypto, err := a.SyncLogin(ctx, client, email, passw, deviceName, deviceType)
		if err != nil {
			return nil, syncclient.FFSyncSession{}, err
		}

		return client, sessionCrypto.Reduce(), nil
	}

	ctx.PrintVerbose(fmt.Sprintf("Sessionfile location is '%s'", ctx.Opt.SessionFilePath))
	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return nil, syncclient.FFSyncSession{}, err
	}

	ctx.PrintVerbose(fmt.Sprintf("Load session from '%s'", cfp))

	if !langext.FileExists(cfp) {
		return nil, syncclient.FFSyncSession{}, fferr.NewDirectOutput(consts.ExitcodeNoLogin, "Sessionfile does not exist.\nUse `ffsclient login <email> <password>` first")
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx, ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		return nil, syncclient.FFSyncSession{}, err
	}

	session, changed, err := client.RefreshSession(ctx, session, ctx.Opt.ForceRefreshSession)
	if err != nil {
		return nil, syncclient.FFSyncSession{}, errorx.Decorate(err, "failed to refresh session")
	}

	if changed && ctx.Opt.SaveRefreshedSession {

		ctx.PrintVerbose("Save new session after auto-update")

		ctx.PrintVerbose("Save session to " + cfp)

		err = session.Save(cfp)
		if err != nil {
			return nil, syncclient.FFSyncSession{}, errorx.Decorate(err, "failed to save session")
		}

	}

	return client, session, nil
}

func (a *CLIArgumentsBaseUtil) SyncLogin(ctx *cli.FFSContext, client *syncclient.FxAClient, email string, password string, devicename string, devicetype string) (syncclient.CryptoSession, error) {

	ctx.PrintVerboseHeader("[1] Login to Sync Account")

	session, err := client.Login(ctx, email, password)
	if err != nil {
		return syncclient.CryptoSession{}, err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[2] Register Device-Name")

	err = client.RegisterDevice(ctx, session, devicename, devicetype)
	if err != nil {
		return syncclient.CryptoSession{}, err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[3] Fetch session keys")

	keyA, keyB, err := client.FetchKeys(ctx, session)
	if err != nil {
		return syncclient.CryptoSession{}, err
	}

	ctx.PrintVerboseKV("Key[a]", keyA)
	ctx.PrintVerboseKV("Key[b]", keyB)

	extsession := session.Extend(keyA, keyB)

	// ========================================================================

	ctx.PrintVerboseHeader("[4] Assert BrowserID")

	sessionBID, err := client.AssertBrowserID(ctx, extsession)
	if err != nil {
		return syncclient.CryptoSession{}, err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[5] Get HAWK Credentials")

	sessionHawk, err := client.HawkAuth(ctx, sessionBID)
	if err != nil {
		return syncclient.CryptoSession{}, err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[6] Get Crypto Keys")

	sessionCrypto, err := client.GetCryptoKeys(ctx, sessionHawk)
	if err != nil {
		return syncclient.CryptoSession{}, err
	}

	return sessionCrypto, nil
}
