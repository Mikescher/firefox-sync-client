package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"regexp"
)

var rexServiceName = regexp.MustCompile(`^[a-zA-Z0-9\-]*$`)

type CLIArgumentsLogin struct {
	Email       string
	Password    string
	ServiceName string
}

func NewCLIArgumentsLogin() *CLIArgumentsLogin {
	return &CLIArgumentsLogin{
		Email:       "",
		Password:    "",
		ServiceName: consts.DefaultServiceName,
	}
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

	a.Email = positionalArgs[0]
	a.Password = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "service-name" && arg.Value != nil {
			a.ServiceName = *arg.Value
			if err := validateServiceName(a.ServiceName); err != nil {
				return errorx.Decorate(err, "invalid service-name")
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
	ctx.PrintVerboseKV("Auth-Server", ctx.Opt.AuthServerURL)
	ctx.PrintVerboseKV("Token-Server", ctx.Opt.TokenServerURL)
	ctx.PrintVerboseKV("Email", a.Email)
	ctx.PrintVerboseKV("Password", a.Password)

	cfp, err := ctx.AbsConfigFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	// ========================================================================

	ctx.PrintVerbose("[1] Login to account")

	session, err := client.Login(ctx, a.Email, a.Password, a.ServiceName)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerbose("[2] Fetch session keys")

	keyA, keyB, err := client.FetchKeys(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerboseKV("Key[a]", keyA)
	ctx.PrintVerboseKV("Key[b]", keyB)

	extsession := session.Extend(keyA, keyB)

	// ========================================================================

	ctx.PrintVerbose("[3] Assert BrowserID")

	sessionBID, err := client.AssertBrowserID(ctx, extsession)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerbose("[4] Get HAWK Credentials")

	sessionHawk, err := client.HawkAuth(ctx, sessionBID)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ctx.PrintVerbose("[4] Get Crypto Keys")

	sessionCrypto, err := client.GetCryptoKeys(ctx, sessionHawk)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	ffsyncSession := sessionCrypto.Reduce()

	ctx.PrintVerbose("Save session-config to " + ctx.Opt.ConfigFilePath)

	err = ffsyncSession.Save(cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Session saved")

	ctx.PrintPrimaryOutput("Succesfully logged in")

	return 0
}

func validateServiceName(name string) error {
	if name == "" {
		return errorx.InternalError.New("Service-name cannot be empty")
	}
	if len(name) > 16 {
		return errorx.InternalError.New("Service-name can be at most 16 characters")
	}
	if !rexServiceName.MatchString(name) {
		return errorx.InternalError.New("Service-name can only contain the characters [A-Z], [a-z], [0-9]")
	}
	return nil
}
