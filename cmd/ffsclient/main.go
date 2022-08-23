package main

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/parser"
	"ffsyncclient/consts"
	"os"
)

func main() {
	verb, opt := parser.ParseCommandline()

	ctx, err := cli.NewContext(opt)
	if err != nil {
		ctx.PrintFatalError(err)
		os.Exit(consts.ExitcodeError)
		return
	}

	defer ctx.Finish()

	exitcode := verb.Execute(ctx)

	os.Exit(exitcode)
}
