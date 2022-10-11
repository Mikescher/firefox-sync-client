package main

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/parser"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"fmt"
	"os"
	"runtime/debug"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("%v\n\n%s", err, string(debug.Stack())))
			os.Exit(consts.ExitcodePanic.Raw)
		}
	}()

	verb, opt, err := parser.ParseCommandline()
	if err != nil {
		ctx := cli.NewEarlyContext()
		ctx.PrintFatalError(err)
		os.Exit(fferr.GetExitCode(err, consts.ExitcodeCLIParse).Raw)
		return
	}

	ctx, err := cli.NewContext(opt)
	if err != nil {
		ctx.PrintFatalError(err)
		os.Exit(fferr.GetExitCode(err, consts.ExitcodeError).Raw)
		return
	}

	defer ctx.Finish()

	err = verb.Execute(ctx)
	if err != nil {
		ctx.PrintFatalError(err)
		os.Exit(fferr.GetExitCode(err, consts.ExitcodeError).Raw)
		return
	}

	os.Exit(consts.ExitcodeOkay.Raw)
}
