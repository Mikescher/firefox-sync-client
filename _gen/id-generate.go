package main

import (
	"gogs.mikescher.com/BlackForestBytes/goext/bfcodegen"
	"os"
)

func main() {
	dest := os.Args[2]

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = bfcodegen.GenerateIDSpecs(wd, dest, bfcodegen.IDGenOptions{})
	if err != nil {
		panic(err)
	}
}
