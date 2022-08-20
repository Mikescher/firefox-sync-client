package cli

import (
	"context"
	"github.com/joomcode/errorx"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type FFSContext struct {
	context.Context
	Opt Options
}

func (c FFSContext) PrintPrimaryOutput(msg string) {
	if c.Opt.Quiet {
		return
	}

	_, err := os.Stdout.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stdout: " + err.Error())
	}
}

func (c FFSContext) PrintFatalMessage(msg string) {
	if c.Opt.Quiet {
		return
	}

	_, err := os.Stderr.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stderr: " + err.Error())
	}
}

func (c FFSContext) PrintFatalError(e error) {
	if c.Opt.Quiet {
		return
	}

	_, err := os.Stderr.WriteString(e.Error() + "\n")
	if err != nil {
		panic("failed to write to stderr: " + err.Error())
	}
}

func (c FFSContext) PrintVerbose(msg string) {
	if c.Opt.Quiet || !c.Opt.Verbose {
		return
	}

	_, err := os.Stdout.WriteString(msg + "\n")
	if err != nil {
		panic("failed to write to stdout: " + err.Error())
	}
}

func (c FFSContext) AbsConfigFilePath() (string, error) {
	fp := c.Opt.ConfigFilePath

	if fp == "~" {
		usr, err := user.Current()
		if err != nil {
			return "", errorx.Decorate(err, "failed to get current user")
		}
		fp = usr.HomeDir
	} else if strings.HasPrefix(fp, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", errorx.Decorate(err, "failed to get current user")
		}
		fp = filepath.Join(usr.HomeDir, fp[2:])
	}

	fp, err := filepath.Abs(fp)
	if err != nil {
		return "", errorx.Decorate(err, "failed to parse config filepath")
	}

	return fp, nil
}

func NewContext(opt Options) *FFSContext {
	return &FFSContext{
		Context: context.Background(),
		Opt:     opt,
	}
}
