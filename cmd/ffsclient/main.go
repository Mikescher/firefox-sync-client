package main

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/parser"
	"os"
)

func main() {
	verb, opt := parser.ParseCommandline()

	exitcode := verb.Execute(cli.NewContext(opt))

	os.Exit(exitcode)
}
