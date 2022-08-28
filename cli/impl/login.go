package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"os"
)

type CLIArgumentsLogin struct {
	Email      string
	Password   string
	DeviceName string
	DeviceType string
}

func NewCLIArgumentsLogin() *CLIArgumentsLogin {
	hostname, err := os.Hostname()
	deviceName := "Firefox-Sync-Client"
	if err == nil {
		deviceName = "Firefox-Sync-Client on " + hostname
	}
	return &CLIArgumentsLogin{
		Email:      "",
		Password:   "",
		DeviceName: deviceName,
	}
}

func (a *CLIArgumentsLogin) Mode() cli.Mode {
	return cli.ModeLogin
}

func (a *CLIArgumentsLogin) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient login <login> <password>", "Login to FF-Sync account, uses ~/.config as default session location"},
		{"          [--device-name=<name>]", ""},
		{"          [--device-type=<type>]", ""},
	}
}

func (a *CLIArgumentsLogin) FullHelp() []string {
	return []string{
		"$> ffsclient login <email> <password> [--device-name] [--device-type]",
		"",
		"Login to FF-Sync account",
		"",
		"This needs to be doe before all other commands that need a valid FirefoxSync connection",
		"If no sesionfile location is provided this uses the default ~/.config/firefox-sync-client.secret",
		"Specify a Device-name to identify the client in the Firefox Account page",
	}
}

func (a *CLIArgumentsLogin) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	if len(positionalArgs) < 2 {
		return errorx.InternalError.New("Not enough arguments for <login>")
	}
	if len(positionalArgs) > 2 {
		return errorx.InternalError.New("Too many arguments for <login>")
	}

	a.Email = positionalArgs[0]
	a.Password = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "device-name" && arg.Value != nil {
			a.DeviceName = *arg.Value
			if err := validateDeviceName(a.DeviceName); err != nil {
				return errorx.Decorate(err, "invalid device-name")
			}
			continue
		}
		if arg.Key == "device-type" && arg.Value != nil {
			a.DeviceType = *arg.Value
			if err := validateDeviceType(a.DeviceType); err != nil {
				return errorx.Decorate(err, "invalid device-type")
			}
			continue
		}
		return errorx.InternalError.New("Unknown argument: " + arg.Key)
	}

	if len(optionArgs) > 0 {
		return errorx.InternalError.New("Unknown argument: " + optionArgs[0].Key)
	}

	return nil
}

func (a *CLIArgumentsLogin) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Login]")
	ctx.PrintVerbose("")

	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)
	ctx.PrintVerboseKV("Email", a.Email)
	ctx.PrintVerboseKV("Password", a.Password)

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	// ========================================================================

	ctx.PrintVerboseHeader("[1] Login to Sync Account")

	session, err := client.Login(ctx, a.Email, a.Password)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[2] Register Device-Name")

	err = client.RegisterDevice(ctx, session, a.DeviceName)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[3] Fetch session keys")

	keyA, keyB, err := client.FetchKeys(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerboseKV("Key[a]", keyA)
	ctx.PrintVerboseKV("Key[b]", keyB)

	extsession := session.Extend(keyA, keyB)

	// ========================================================================

	ctx.PrintVerboseHeader("[4] Assert BrowserID")

	sessionBID, err := client.AssertBrowserID(ctx, extsession)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[5] Get HAWK Credentials")

	sessionHawk, err := client.HawkAuth(ctx, sessionBID)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[6] Get Crypto Keys")

	sessionCrypto, err := client.GetCryptoKeys(ctx, sessionHawk)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[7] Save Session")

	ffsyncSession := sessionCrypto.Reduce()

	ctx.PrintVerbose("Save session to " + ctx.Opt.SessionFilePath)

	err = ffsyncSession.Save(cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Session saved")

	ctx.PrintPrimaryOutput("Succesfully logged in")

	return 0
}

func validateDeviceName(name string) error {
	if name == "" {
		return errorx.InternalError.New("Device-name cannot be empty")
	}
	if len(name) > 255 {
		return errorx.InternalError.New("Device-name can be at most 16 characters")
	}
	return nil
}

func validateDeviceType(name string) error {
	if name == "" {
		return errorx.InternalError.New("Device-type cannot be empty")
	}
	if len(name) > 16 {
		return errorx.InternalError.New("Device-type can be at most 16 characters")
	}
	return nil
}
