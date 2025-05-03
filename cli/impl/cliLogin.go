package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"os"
)

type CLIArgumentsLogin struct {
	Email      string
	Password   string
	DeviceName string
	DeviceType string
	CLIArgumentsBaseUtil
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
		DeviceType: "cli",
	}
}

func (a *CLIArgumentsLogin) Mode() cli.Mode {
	return cli.ModeLogin
}

func (a *CLIArgumentsLogin) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsLogin) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsLogin) ShortHelp() [][]string {

	// --otp is theoretically a global option, but kinda behaves like an option of cliLogin
	// because its only useful globally in combination with --auth-login-*

	return [][]string{
		{"ffsclient login <login> <password>", "Login to FF-Sync account, uses ~/.config as default session location"},
		{"          [--device-name=<name>]", "Send your device-name to identify the session later"},
		{"          [--device-type=<type>]", "Send your device-type to identify the session later"},
		{"          [--otp=<value>]", "A valid TOTP token, in case one is needed for the login"},
	}
}

func (a *CLIArgumentsLogin) FullHelp() []string {
	return []string{
		"$> ffsclient login <email> <password> [--device-name] [--device-type]",
		"",
		"Login to FF-Sync account",
		"",
		"This needs to be done before all other commands that need a valid FirefoxSync connection",
		"If no sesionfile location is provided this uses the default ~/.config/firefox-sync-client.secret",
		"Specify a Device-name to identify the client in the Firefox Account page",
		"if a 2-Factor TOTP token is needed an prompt will request one, or a totp can be pre-supplied with the --otp parameter",
	}
}

func (a *CLIArgumentsLogin) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
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
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsLogin) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Login]")
	ctx.PrintVerbose("")

	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)
	ctx.PrintVerboseKV("Email", a.Email)
	ctx.PrintVerboseKV("Password", a.Password)

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		return err
	}

	client := syncclient.NewFxAClient(ctx, ctx.Opt.AuthServerURL)

	// ========================================================================

	sessionCrypto, err := a.SyncLogin(ctx, client, a.Email, a.Password, a.DeviceName, a.DeviceType)
	if err != nil {
		return err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[7] Save Session")

	ffsyncSession := sessionCrypto.Reduce()

	ctx.PrintVerbose("Save session to " + ctx.Opt.SessionFilePath)

	err = ffsyncSession.Save(cfp)
	if err != nil {
		return err
	}

	ctx.PrintVerbose("Session saved")

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput("Succesfully logged in")

	return nil
}

func validateDeviceName(name string) error {
	if name == "" {
		return fferr.DirectOutput.New("Device-name cannot be empty")
	}
	if len(name) > 255 {
		return fferr.DirectOutput.New("Device-name can be at most 16 characters")
	}
	return nil
}

func validateDeviceType(name string) error {
	if name == "" {
		return fferr.DirectOutput.New("Device-type cannot be empty")
	}
	if len(name) > 16 {
		return fferr.DirectOutput.New("Device-type can be at most 16 characters")
	}
	return nil
}
